package template_router

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/crop_model"
	"armazenda/model/field_model"
	"armazenda/model/person_model"
	"armazenda/model/vehicle_model"
	"armazenda/service/user_service"
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TemplateFS holds the embedded templates filesystem
var templateFS fs.FS
var htmlTemplate *template.Template

// InitTemplateRouter initializes the template router with the filesystem and compiled templates
func InitTemplateRouter(fs fs.FS, tmpl *template.Template) {
	templateFS = fs
	htmlTemplate = tmpl
}

// serveTemplate serves a template with optional pre-rendering for offline use
func serveTemplate(c *gin.Context) {
	name := c.Param("name")

	// Security: prevent directory traversal
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		c.String(http.StatusBadRequest, "Invalid template name")
		return
	}

	// Check if this is a form template that needs pre-rendering
	if isPreRenderTemplate(name) {
		html, err := servePreRenderedTemplate(c, name)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to render template: %v", err))
			return
		}
		setCacheHeaders(c)
		c.String(http.StatusOK, html)
		return
	}

	// For list items and other templates, serve raw Go template
	content, err := findTemplate(name)
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Template not found: %s", name))
		return
	}

	setCacheHeaders(c)
	c.String(http.StatusOK, string(content))
}

// isPreRenderTemplate returns true if the template should be pre-rendered with reference data
func isPreRenderTemplate(name string) bool {
	preRenderTemplates := map[string]bool{
		"entry-form":           true,
		"entry-draft-form":     true,
		"departure-form":       true,
		"departure-draft-form": true,
		"person-form":          true,
		"entry-list-item":      true,
		"departure-list-item":  true,
		"person-list-item":     true,
	}
	return preRenderTemplates[name]
}

// servePreRenderedTemplate renders a template with reference data and transforms it for client-side use
func servePreRenderedTemplate(c *gin.Context, name string) (string, error) {
	// Check if this is a list item template that needs raw template transformation
	if strings.Contains(name, "list-item") {
		return serveListItemTemplate(name)
	}

	// Get farm from session
	sid, _ := c.Cookie("session_id")
	farm := user_service.GetFarmFromToken(sid)

	// Prepare data for template rendering
	data, err := prepareTemplateData(name, farm)
	if err != nil {
		return "", fmt.Errorf("failed to prepare data: %w", err)
	}

	// Execute template with data
	var buf bytes.Buffer
	if err := htmlTemplate.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Transform the rendered HTML for client-side placeholder replacement
	html := transformTemplateForClient(buf.String(), name)

	return html, nil
}

// serveListItemTemplate reads raw template and transforms Go template syntax to client placeholders
func serveListItemTemplate(name string) (string, error) {
	// Read the raw template content
	content, err := findTemplate(name)
	if err != nil {
		return "", fmt.Errorf("failed to find template: %w", err)
	}

	// Transform Go template syntax {{ .Field }} to client placeholders {Field}
	html := transformListItemTemplate(string(content))

	return html, nil
}

// transformListItemTemplate converts Go template syntax to client-side placeholders
func transformListItemTemplate(template string) string {
	result := template

	// Convert {{ .Id }} and {{ .Id}} (with/without space) to {Id}
	result = regexp.MustCompile(`\{\{\s*\.Id\s*\}\}`).ReplaceAllString(result, "{Id}")
	// Convert {{ .Product }} to {Product}
	result = regexp.MustCompile(`\{\{\s*\.Product\s*\}\}`).ReplaceAllString(result, "{Product}")
	// Convert {{ .Origin }} to {Origin}
	result = regexp.MustCompile(`\{\{\s*\.Origin\s*\}\}`).ReplaceAllString(result, "{Origin}")
	// Convert {{ .Field }} to {Field}
	result = regexp.MustCompile(`\{\{\s*\.Field\s*\}\}`).ReplaceAllString(result, "{Field}")
	// Convert {{ .Vehicle }} to {Vehicle}
	result = regexp.MustCompile(`\{\{\s*\.Vehicle\s*\}\}`).ReplaceAllString(result, "{Vehicle}")
	// Convert {{ .NetWeight }} to {NetWeight}
	result = regexp.MustCompile(`\{\{\s*\.NetWeight\s*\}\}`).ReplaceAllString(result, "{NetWeight}")
	// Convert {{ .ArrivalDate }} to {ArrivalDate}
	result = regexp.MustCompile(`\{\{\s*\.ArrivalDate\s*\}\}`).ReplaceAllString(result, "{ArrivalDate}")
	// Convert {{ .DepartureDate }} to {DepartureDate}
	result = regexp.MustCompile(`\{\{\s*\.DepartureDate\s*\}\}`).ReplaceAllString(result, "{DepartureDate}")
	// Convert {{ .GrossWeight }} to {GrossWeight}
	result = regexp.MustCompile(`\{\{\s*\.GrossWeight\s*\}\}`).ReplaceAllString(result, "{GrossWeight}")
	// Convert {{ .Tare }} to {Tare}
	result = regexp.MustCompile(`\{\{\s*\.Tare\s*\}\}`).ReplaceAllString(result, "{Tare}")
	// Convert {{ .Humidity }} to {Humidity}
	result = regexp.MustCompile(`\{\{\s*\.Humidity\s*\}\}`).ReplaceAllString(result, "{Humidity}")
	// Convert {{ .Damage }} to {Damage}
	result = regexp.MustCompile(`\{\{\s*\.Damage\s*\}\}`).ReplaceAllString(result, "{Damage}")
	// Convert {{ .Impurity }} to {Impurity}
	result = regexp.MustCompile(`\{\{\s*\.Impurity\s*\}\}`).ReplaceAllString(result, "{Impurity}")

	return result
}

// prepareTemplateData prepares the data structure needed for template rendering
func prepareTemplateData(name string, farm uint32) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	switch name {
	case "entry-form", "entry-draft-form":
		// Fetch reference data
		fields, fieldsErr := field_model.GetFieldModel().GetFieldsByFarm(farm)
		if fieldsErr != nil {
			return nil, fieldsErr
		}
		data["Fields"] = fields

		crops, cropsErr := crop_model.GetCropModel().GetCropsByFarm(farm)
		if cropsErr != nil {
			return nil, cropsErr
		}
		data["Crops"] = crops

		vehicleModel, vmErr := vehicle_model.GetVehicleModel()
		if vmErr != nil {
			return nil, vmErr
		}
		vehicles, vehiclesErr := vehicleModel.GetVehiclesByFarm(farm)
		if vehiclesErr != nil {
			return nil, vehiclesErr
		}
		data["Vehicles"] = vehicles

		people, peopleErr := person_model.GetPersonModel().GetPeopleByFarm(farm)
		if peopleErr != nil {
			return nil, peopleErr
		}
		data["People"] = people

		// Empty entry for new form
		data["Entry"] = entity_public.Entry{}
		data["IsOffline"] = false

	case "departure-form", "departure-draft-form":
		// Fetch reference data
		crops, cropsErr := crop_model.GetCropModel().GetCropsByFarm(farm)
		if cropsErr != nil {
			return nil, cropsErr
		}
		data["Crops"] = crops

		vehicleModel, vmErr := vehicle_model.GetVehicleModel()
		if vmErr != nil {
			return nil, vmErr
		}
		vehicles, vehiclesErr := vehicleModel.GetVehiclesByFarm(farm)
		if vehiclesErr != nil {
			return nil, vehiclesErr
		}
		data["Vehicles"] = vehicles

		people, peopleErr := person_model.GetPersonModel().GetPeopleByFarm(farm)
		if peopleErr != nil {
			return nil, peopleErr
		}
		data["People"] = people

		// Empty departure for new form
		data["Departure"] = entity_public.Departure{}
		data["IsOffline"] = false

	case "person-form":
		// Person form doesn't need reference data
		data["IsOffline"] = false
	}

	return data, nil
}

// transformTemplateForClient transforms rendered HTML to use placeholders for client-side editing
func transformTemplateForClient(html string, templateName string) string {
	result := html

	// Transform entry-specific fields
	if strings.Contains(templateName, "entry") {
		result = transformEntryForm(result)
	} else if strings.Contains(templateName, "departure") {
		result = transformDepartureForm(result)
	}

	// Transform common patterns for all templates
	result = transformCommonPatterns(result)

	return result
}

// transformEntryForm transforms entry form HTML
func transformEntryForm(html string) string {
	result := html

	// Replace Entry.Id with placeholder for conditional rendering
	// Find the h1 element with conditional content
	h1Pattern := regexp.MustCompile(`(?s)(<h1[^>]*>)(.+?)(</h1>)`)
	result = h1Pattern.ReplaceAllStringFunc(result, func(match string) string {
		// Check if it contains both "Editando" and "Nova"
		if strings.Contains(match, "Editando") && strings.Contains(match, "Nova") {
			// Replace with conditional markup
			return `<h1 data-show-if="Entry.Id">Editando Entrada</h1><h1 data-hide-if="Entry.Id">Nova Entrada</h1>`
		}
		return match
	})

	// Replace specific input values with placeholders
	result = replaceInputValueByName(result, "grossWeight", "Entry.GrossWeight")
	result = replaceInputValueByName(result, "tare", "Entry.Tare")
	result = replaceInputValueByName(result, "humidity", "Entry.Humidity")
	result = replaceInputValueByName(result, "damage", "Entry.Damage")
	result = replaceInputValueByName(result, "impurity", "Entry.Impurity")
	result = replaceInputValueByName(result, "arrivalDate", "Entry.ArrivalDate")
	result = replaceInputValueByName(result, "netWeightRaw", "Entry.NetWeight")
	result = replaceInputValueByName(result, "netWeight", "Entry.NetWeight")

	// Entry.Id in hx-put attribute
	result = regexp.MustCompile(`hx-put="/entry/[^"]*"`).ReplaceAllString(result, `hx-put="/entry/{Entry.Id}"`)
	result = regexp.MustCompile(`hx-target="#entry-[^"]*"`).ReplaceAllString(result, `hx-target="#entry-{Entry.Id}"`)

	return result
}

// transformDepartureForm transforms departure form HTML
func transformDepartureForm(html string) string {
	result := html

	// Replace h1 conditional
	h1Pattern := regexp.MustCompile(`(?s)(<h1[^>]*>)(.+?)(</h1>)`)
	result = h1Pattern.ReplaceAllStringFunc(result, func(match string) string {
		if strings.Contains(match, "Editando") && strings.Contains(match, "Nova") {
			return `<h1 data-show-if="Departure.Id">Editando Saída</h1><h1 data-hide-if="Departure.Id">Nova Saída</h1>`
		}
		return match
	})

	// Replace input values using targeted string replacement
	result = replaceInputValueByName(result, "grossWeight", "Departure.GrossWeight")
	result = replaceInputValueByName(result, "tare", "Departure.Tare")
	result = replaceInputValueByName(result, "humidity", "Departure.Humidity")
	result = replaceInputValueByName(result, "damage", "Departure.Damage")
	result = replaceInputValueByName(result, "impurity", "Departure.Impurity")
	result = replaceInputValueByName(result, "departureDate", "Departure.DepartureDate")
	result = replaceInputValueByName(result, "netWeight", "Departure.NetWeight")

	result = regexp.MustCompile(`hx-put="/departure/[^"]*"`).ReplaceAllString(result, `hx-put="/departure/{Departure.Id}"`)

	return result
}

// replaceInputValueByName finds an input element by name attribute and replaces its value attribute
func replaceInputValueByName(html, inputName, placeholder string) string {
	// Find name="inputName"
	nameIdx := strings.Index(html, `name="`+inputName+`"`)
	if nameIdx == -1 {
		return html
	}

	// Find opening < of this tag (search backward from nameIdx)
	tagStart := strings.LastIndex(html[:nameIdx], "<")
	if tagStart == -1 {
		return html
	}

	// Find closing > of this tag (search forward from nameIdx)
	tagEnd := strings.Index(html[nameIdx:], ">")
	if tagEnd == -1 {
		return html
	}
	tagEnd += nameIdx

	// Now search for value=" within the full tag range
	tagContent := html[tagStart:tagEnd]
	valueStart := strings.Index(tagContent, `value="`)
	if valueStart == -1 {
		return html
	}

	// Calculate absolute positions in original html
	// valueOpenPos: position right after value="
	valueOpenPos := tagStart + valueStart + 7
	// Find closing quote after value="
	closingQuotePos := strings.Index(html[valueOpenPos:], `"`)
	if closingQuotePos == -1 {
		return html
	}
	// valueClosePos: absolute position of the closing quote
	valueClosePos := valueOpenPos + closingQuotePos

	// Replace the value content (between quotes)
	return html[:valueOpenPos] + "{" + placeholder + "}" + html[valueClosePos:]
}

// transformCommonPatterns transforms patterns common to all templates
func transformCommonPatterns(html string) string {
	result := html

	// Replace button text conditionals
	buttonPattern := regexp.MustCompile(`(?s)<button[^>]*type="submit"[^>]*>(.+?)</button>`)
	result = buttonPattern.ReplaceAllStringFunc(result, func(match string) string {
		if strings.Contains(match, "Salvar") && strings.Contains(match, "Adicionar") {
			// Keep the button with both options, client will show/hide
			return `<button class="primary-glass-btn" type="submit"><span data-show-if="Entry.Id">Salvar</span><span data-hide-if="Entry.Id">Adicionar</span></button>`
		}
		return match
	})

	// Transform script tags that call setupEntryForm with server-side data
	// Replace the argument with a placeholder for client-side replacement
	result = regexp.MustCompile(`setupEntryForm\([^)]*\)`).ReplaceAllString(result, `setupEntryForm({Entry.ArrivalDate})`)

	return result
}

// setCacheHeaders sets 24-hour cache headers
func setCacheHeaders(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("Expires", time.Now().Add(24*time.Hour).Format(http.TimeFormat))
	c.Header("Content-Type", "text/html; charset=utf-8")
}

// findTemplate searches for a template file by name in the templates directory
func findTemplate(name string) ([]byte, error) {
	// Try direct path first
	path := fmt.Sprintf("templates/%s.html", name)
	content, err := fs.ReadFile(templateFS, path)
	if err == nil {
		return content, nil
	}

	// Walk through templates directory to find the file
	return walkTemplates(name)
}

// walkTemplates walks through the templates directory to find a template by name
func walkTemplates(name string) ([]byte, error) {
	var result []byte

	err := fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Check if this is the file we're looking for
		base := filepath.Base(path)
		expected := fmt.Sprintf("%s.html", name)
		if base == expected {
			content, readErr := fs.ReadFile(templateFS, path)
			if readErr == nil {
				result = content
				return fs.SkipAll // Stop walking once found
			}
		}

		return nil
	})

	if err != nil && err != fs.SkipAll {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	return result, nil
}

// UseTemplateRoutes sets up template routes
func UseTemplateRoutes(router *gin.Engine) {
	router.GET("/api/templates/:name", serveTemplate)
}
