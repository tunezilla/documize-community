// Copyright 2016 Documize Inc. <legal@documize.com>. All rights reserved.
//
// This software (Documize Community Edition) is licensed under
// GNU AGPL v3 http://www.gnu.org/licenses/agpl-3.0.en.html
//
// You can operate outside the AGPL restrictions by purchasing
// Documize Enterprise Edition and obtaining a commercial license
// by contacting <sales@documize.com>.
//
// https://documize.com

package markdown

import (
	"net/http"

	"github.com/documize/blackfriday"
	"github.com/documize/community/core/env"
	"github.com/documize/community/domain/section/provider"
	"github.com/documize/community/domain/store"
	"github.com/microcosm-cc/bluemonday"
)

// Provider represents Markdown
type Provider struct {
	Runtime *env.Runtime
	Store   *store.Store
}

// Meta describes us
func (*Provider) Meta() provider.TypeMeta {
	section := provider.TypeMeta{}

	section.ID = "1470bb4a-36c6-4a98-a443-096f5658378b"
	section.Title = "Markdown"
	section.Description = "CommonMark based content"
	section.ContentType = "markdown"
	section.PageType = "section"
	section.Order = 9998

	return section
}

// Command stub.
func (*Provider) Command(ctx *provider.Context, w http.ResponseWriter, r *http.Request) {
	provider.WriteEmpty(w)
}

// Render converts markdown data into HTML suitable for browser rendering.
// See also https://github.com/yuin/goldmark
func (*Provider) Render(ctx *provider.Context, config, data string) string {
	unsafe := blackfriday.Run([]byte(data))

	policy := bluemonday.UGCPolicy()
	policy.AllowDataURIImages()

	safe := policy.SanitizeBytes(unsafe)
	return string(safe)
}

// Refresh just sends back data as-is.
func (*Provider) Refresh(ctx *provider.Context, config, data string) string {
	return data
}
