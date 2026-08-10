package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"math/rand"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	uuid "github.com/nu7hatch/gouuid"

	"github.com/documize/community/core/secrets"
	"github.com/documize/community/core/uniqueid"
	"github.com/documize/community/domain/attachment"
	"github.com/documize/community/model/page"
	mat "github.com/documize/community/model/attachment"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
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
		log.Printf("could not parse page #%d: %v", page.ID, err)
		return
	}

	changed := false
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.Img {
			for i := range n.Attr {
				a := n.Attr[i]
				if a.Key == "src" {
					if strings.HasPrefix(a.Val, "data:") {
						location, err := attachmentize(db, storer, page, a.Val)
						if err != nil {
							log.Printf("could not create blob: %v", err)
						} else {
							changed = true
							n.Attr[i].Val = location
						}
					}
				}
			}
		}
	}

	if changed {
		var buf strings.Builder
		html.Render(&buf, doc)
		previousBodySize := len(page.Body)
		page.Body = buf.String()

		_, err = db.Exec("UPDATE dmz_section SET c_body = $1 WHERE id = $2", page.Body, page.ID)
		if err != nil {
			log.Printf("could not update page with new body")
		}

		log.Printf("previousPageLen=%d currentPageLen=%d", previousBodySize, len(page.Body))
	}
}

func contentTypeToExtension(contentType string) string {
	if strings.HasPrefix(contentType, "image/") {
		after, _ := strings.CutPrefix(contentType, "image/")
		return after
	}

	return "txt"
}

func attachmentize(db *sql.DB, storer attachment.ObjectStorer, page *page.Page, imgBlob string) (string, error) {
	blob, contentType, err := dataURIToBlob(imgBlob)
	if err != nil {
		return "", err
	}

	newUUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	extension := contentTypeToExtension(contentType)
	fileName := randStringBytes(12)
	random := secrets.GenerateSalt()

	a := mat.Attachment{
		OrgID: page.OrgID,
		DocumentID: page.DocumentID,
		SectionID: page.RefID,
		Job: newUUID.String(),
		FileID: random[0:9],
		Filename: fileName,
		Data: []byte(""),
		Extension: extension,
	}
	a.RefID = uniqueid.Generate()
	a.Created = time.Now().UTC()
	a.Revised = time.Now().UTC()

	err = storer.PutNoContext(a, blob)
	if err != nil {
		return "", err
	}

	_, err = db.Exec("INSERT INTO dmz_doc_attachment (c_refid, c_orgid, c_docid, c_sectionid, c_job, c_fileid, c_filename, c_data, c_extension, c_created, c_revised) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", a.RefID, a.OrgID, a.DocumentID, a.SectionID, a.Job, a.FileID, a.Filename, a.Data, a.Extension, a.Created, a.Revised)
	if err != nil {
		return "", err
	}

	location := fmt.Sprintf("/api/public/attachment/%s/%s", a.OrgID, a.RefID)
	return location, nil
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
	if minioBucket == "" {
		minioBucket = "community"
	}
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
