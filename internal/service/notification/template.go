package notification

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	htmltemplate "html/template"
	"sync"
	texttemplate "text/template"
)

// EmailTemplateOverride mirrors the per-namespace override stored on
// NamespaceConfig. Empty fields fall back to the registered template.
type EmailTemplateOverride struct {
	Subject  string
	HTMLBody string
	TextBody string
}

// SMSTemplateOverride mirrors the per-namespace SMS override.
type SMSTemplateOverride struct {
	Body string
}

// EmailTemplate is a code-defined template registered at startup.
type EmailTemplate struct {
	Name    string
	Subject *texttemplate.Template
	HTML    *htmltemplate.Template
	Text    *texttemplate.Template
}

// SMSTemplate is a code-defined SMS template.
type SMSTemplate struct {
	Name string
	Body *texttemplate.Template
}

// MustEmailTemplate parses and panics on error. Use at package init time.
func MustEmailTemplate(name, subject, htmlBody, textBody string) EmailTemplate {
	t := EmailTemplate{Name: name}
	t.Subject = texttemplate.Must(texttemplate.New(name + ":subject").Parse(subject))
	t.HTML = htmltemplate.Must(htmltemplate.New(name + ":html").Parse(htmlBody))
	if textBody != "" {
		t.Text = texttemplate.Must(texttemplate.New(name + ":text").Parse(textBody))
	}
	return t
}

// MustSMSTemplate parses and panics on error.
func MustSMSTemplate(name, body string) SMSTemplate {
	return SMSTemplate{
		Name: name,
		Body: texttemplate.Must(texttemplate.New(name + ":body").Parse(body)),
	}
}

// RenderedEmail is the output of template rendering.
type RenderedEmail struct {
	Subject  string
	HTMLBody string
	TextBody string
}

// TemplateRegistry holds registered code templates and a cache of parsed overrides.
type TemplateRegistry struct {
	mu            sync.RWMutex
	emails        map[string]EmailTemplate
	sms           map[string]SMSTemplate
	textOverrides sync.Map // map[string]*text/template.Template
	htmlOverrides sync.Map // map[string]*html/template.Template
}

func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		emails: make(map[string]EmailTemplate),
		sms:    make(map[string]SMSTemplate),
	}
}

func (r *TemplateRegistry) RegisterEmail(t EmailTemplate) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.emails[t.Name] = t
}

func (r *TemplateRegistry) RegisterSMS(t SMSTemplate) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sms[t.Name] = t
}

func (r *TemplateRegistry) RenderEmail(namespace, name string, override *EmailTemplateOverride, data any) (RenderedEmail, error) {
	r.mu.RLock()
	base, ok := r.emails[name]
	r.mu.RUnlock()
	if !ok {
		return RenderedEmail{}, fmt.Errorf("%w: email %q", ErrTemplateNotFound, name)
	}

	out := RenderedEmail{}

	subj := base.Subject
	if override != nil && override.Subject != "" {
		t, err := r.parsedTextOverride(namespace, name, "subject", override.Subject)
		if err != nil {
			return RenderedEmail{}, err
		}
		subj = t
	}
	subjStr, err := execText(subj, data)
	if err != nil {
		return RenderedEmail{}, err
	}
	out.Subject = subjStr

	htmlT := base.HTML
	if override != nil && override.HTMLBody != "" {
		t, err := r.parsedHTMLOverride(namespace, name, "html", override.HTMLBody)
		if err != nil {
			return RenderedEmail{}, err
		}
		htmlT = t
	}
	htmlStr, err := execHTML(htmlT, data)
	if err != nil {
		return RenderedEmail{}, err
	}
	out.HTMLBody = htmlStr

	var textT *texttemplate.Template
	if override != nil && override.TextBody != "" {
		t, err := r.parsedTextOverride(namespace, name, "text", override.TextBody)
		if err != nil {
			return RenderedEmail{}, err
		}
		textT = t
	} else if base.Text != nil {
		textT = base.Text
	}
	if textT != nil {
		textStr, err := execText(textT, data)
		if err != nil {
			return RenderedEmail{}, err
		}
		out.TextBody = textStr
	}

	return out, nil
}

func (r *TemplateRegistry) RenderSMS(namespace, name string, override *SMSTemplateOverride, data any) (string, error) {
	r.mu.RLock()
	base, ok := r.sms[name]
	r.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("%w: sms %q", ErrTemplateNotFound, name)
	}
	bodyT := base.Body
	if override != nil && override.Body != "" {
		t, err := r.parsedTextOverride(namespace, name, "body", override.Body)
		if err != nil {
			return "", err
		}
		bodyT = t
	}
	return execText(bodyT, data)
}

func (r *TemplateRegistry) parsedTextOverride(ns, name, field, src string) (*texttemplate.Template, error) {
	key := overrideKey(ns, name, field, src)
	if v, ok := r.textOverrides.Load(key); ok {
		return v.(*texttemplate.Template), nil
	}
	t, err := texttemplate.New(name + ":" + field).Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parse override %s/%s/%s: %w", ns, name, field, err)
	}
	actual, _ := r.textOverrides.LoadOrStore(key, t)
	return actual.(*texttemplate.Template), nil
}

func (r *TemplateRegistry) parsedHTMLOverride(ns, name, field, src string) (*htmltemplate.Template, error) {
	key := overrideKey(ns, name, field, src)
	if v, ok := r.htmlOverrides.Load(key); ok {
		return v.(*htmltemplate.Template), nil
	}
	t, err := htmltemplate.New(name + ":" + field).Parse(src)
	if err != nil {
		return nil, fmt.Errorf("parse override %s/%s/%s: %w", ns, name, field, err)
	}
	actual, _ := r.htmlOverrides.LoadOrStore(key, t)
	return actual.(*htmltemplate.Template), nil
}

func overrideKey(ns, name, field, src string) string {
	sum := sha256.Sum256([]byte(src))
	return ns + "|" + name + "|" + field + "|" + hex.EncodeToString(sum[:8])
}

func execText(t *texttemplate.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func execHTML(t *htmltemplate.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
