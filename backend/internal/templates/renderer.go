package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed email/*.html
var emailFS embed.FS

// Data must be implemented by all email template data structs.
type Data interface {
	GetSubject() string
	GetAppName() string
}

// Base holds the fields common to every email template data struct.
// Embed this in all concrete data types.
type Base struct {
	AppName string
	Subject string
}

func (b Base) GetSubject() string { return b.Subject }
func (b Base) GetAppName() string { return b.AppName }

// Renderer renders named email templates wrapped in the shared layout.
type Renderer struct {
	layout *template.Template
}

func NewRenderer() (*Renderer, error) {
	layout, err := template.ParseFS(emailFS, "email/layout.html")
	if err != nil {
		return nil, fmt.Errorf("parse email layout: %w", err)
	}
	return &Renderer{layout: layout}, nil
}

// Render renders the named content template with data, wrapping it in the
// layout. Returns the email subject and full HTML body.
func (r *Renderer) Render(name string, data Data) (subject, html string, err error) {
	// Render content template to buffer
	contentTmpl, err := template.ParseFS(emailFS, "email/"+name+".html")
	if err != nil {
		return "", "", fmt.Errorf("parse content template %q: %w", name, err)
	}
	var contentBuf bytes.Buffer
	if err = contentTmpl.Execute(&contentBuf, data); err != nil {
		return "", "", fmt.Errorf("execute content template %q: %w", name, err)
	}

	// Render layout with rendered content
	layoutData := struct {
		AppName string
		Subject string
		Content template.HTML
	}{
		AppName: data.GetAppName(),
		Subject: data.GetSubject(),
		Content: template.HTML(contentBuf.String()),
	}
	var htmlBuf bytes.Buffer
	if err = r.layout.Execute(&htmlBuf, layoutData); err != nil {
		return "", "", fmt.Errorf("execute layout for %q: %w", name, err)
	}
	return data.GetSubject(), htmlBuf.String(), nil
}
