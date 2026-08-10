package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"path"

	_ "github.com/lib/pq"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/documize/community/domain/attachment"
	"github.com/documize/community/domain"
	"github.com/documize/community/model/page"
	mat "github.com/documize/community/model/attachment"
)

func attachmentS3Path(orgID string, refID string, fileName string) string {
	return fmt.Sprintf("%v/%v/%v", orgID, refID, path.Clean(fileName))
}

func migratePages(db *sql.DB, storer attachment.ObjectStorer) {
	const q = `
		SELECT id, c_refid, c_orgid, c_docid, c_userid, c_contenttype, c_body
		FROM dmz_section
		`

	rows, err := db.Query(q)
	if err != nil {
		log.Fatalf("couldn't query for pages: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p page.Page
		if err := rows.Scan(&p.ID, &p.RefID, &p.OrgID, &p.DocumentID, &p.UserID, &p.ContentType, &p.Body); err != nil {
			log.Printf("couldn't get next page: %v", err)
		}

		processPage(db, &p, storer)
	}

	err = rows.Err()
	if err != nil {
		log.Printf("error during row iteration: %v", err)
	}
}

// replaces body content of the page and uploads the images
func processPage(db *sql.DB, page *page.Page, storer attachment.ObjectStorer) {
	doc, err := html.Parse(strings.NewReader(page.Body))
	if err != nil {
		log.Printf("could not parse page #%i: %v", page.ID, err)
		return
	}

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.Img {
			for _, a := range n.Attr {
				if a.Key == "src" {
					if strings.HasPrefix(a.Val, "data:") {
						uri, err := attachmentize(db, storer, page, a.Val)
						if err != nil {
							log.Printf("could not create blob: %v", err)
						} else {
							a.Val = uri
						}
					}
				}
			}
		}
	}

	var buf strings.Builder
	html.Render(&buf, doc)
	page.Body = buf.String()
}

func attachmentize(db *sql.DB, storer attachment.ObjectStorer, page *page.Page, imgBlob string) (string, error) {
	blob, contentType, err := dataURIToBlob(imgBlob)
	if err != nil {
		return "", err
	}

	a := mat.Attachment{
		OrgID: page.OrgID,
		DocumentID: page.DocumentID,
		SectionID: page.ID,
		Job: "",
		FileID: "",
		Filename: "",
		Data: []byte(""),
		Extension: "",
	}
}

func dataURIToBlob(uri string) (data []byte, contentType string, err error) {
	if !strings.HasPrefix(uri, "data:") {
		return nil, "", fmt.Errorf("not a data URI")
	}

	comma := strings.IndexByte(uri, ',')
	if comma == -1 {
		return nil, "", fmt.Errorf("malformed data URI: no comma")
	}

	meta := uri[5:comma]
	payload := uri[comma+1:]
	isBase64 := false
	if i := strings.IndexByte(meta, ';'); i != -1 {
		contentType = meta[:i]
		if strings.Contains(meta[i:], "base64") {
			isBase64 = true
		}
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if isBase64 {
		data, err = base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return nil, "", fmt.Errorf("base64 decode: %w", err)
		}
	} else {
		data = []byte(payload)
	}

	return data, contentType, nil
}

func main() {
	var (
		objectStorer attachment.ObjectStorer
		err          error
	)

	minioUrl := os.Getenv("DOCUMIZEMINIO")
	minioBucket := os.Getenv("DOCUMIZEMINIOBUCKET")
	if minioUrl == "" {
		objectStorer, err = attachment.S3Storer(minioBucket)
	} else {
		objectStorer, err = attachment.MinioStorer(minioBucket, minioUrl)
	}

	dbSettings := os.Getenv("DOCUMIZEDB")
	db, err := sql.Open("postgres", dbSettings)
	if err != nil {
		log.Fatalf("couldn't open pg: %v", err)
	}

	migratePages(db, objectStorer)
}
