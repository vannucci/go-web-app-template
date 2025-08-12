package destinations

import (
    "encoding/json"
    "html/template"
    "net/http"
    "strconv"
    "strings"
    "time"
    
    "github.com/gorilla/mux"
    "main-server/handlers"
    "main-server/models/destinations"
)

type MetaHandler struct {
    *handlers.BaseHandler
}

func NewMetaHandler(deps *handlers.Dependencies) *MetaHandler {
    return &MetaHandler{
        BaseHandler: handlers.NewBaseHandler(deps),
    }
}

func (h *MetaHandler) ShowForm(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    
    var dest destinations.MetaDestination
    dest.Type = "meta"
    
    // Load existing if editing
    if id := vars["id"]; id != "" {
        destID, _ := strconv.Atoi(id)
        if err := h.LoadDestination(destID, &dest); err != nil {
            http.Error(w, "Destination not found", http.StatusNotFound)
            return
        }
    }
    
    data := map[string]interface{}{
        "Title":       "Meta Destination Configuration",
        "Type":        "meta",
        "Destination": dest,
        "MinDate":     time.Now().Format("2006-01-02"),
        "Errors":      make(map[string]string),
    }
    
    h.RenderTemplate(w, "destinations/meta.html", data)
}

func (h *MetaHandler) Validate(w http.ResponseWriter, r *http.Request) {
    dest := h.parseMetaForm(r)
    
    w.Header().Set("Content-Type", "text/html")
    
    if err := dest.Validate(); err != nil {
        w.Write([]byte(`<div class="error">` + template.HTMLEscapeString(err.Error()) + `</div>`))
        return
    }
    
    w.Write([]byte(`<div class="validation-live">✓ All fields valid</div>`))
}

func (h *MetaHandler) Save(w http.ResponseWriter, r *http.Request) {
    dest := h.parseMetaForm(r)
    
    // Validate
    if err := dest.Validate(); err != nil {
        // Re-render form with errors
        data := map[string]interface{}{
            "Title":       "Meta Destination Configuration",
            "Type":        "meta",
            "Destination": dest,
            "MinDate":     time.Now().Format("2006-01-02"),
            "Error":       err.Error(),
        }
        h.RenderTemplate(w, "destinations/meta.html", data)
        return
    }
    
    // Get current user from context
    user := h.GetCurrentUser(r)
    dest.CompanyID = user.CompanyID
    dest.CreatedBy = user.ID
    
    // Convert to JSON for storage
    configJSON, _ := json.Marshal(dest)
    
    // Save to database
    query := `
        INSERT INTO destinations (company_id, workspace_id, platform_id, name, type, config, created_by)
        SELECT $1, c.workspace_id, p.id, $2, $3, $4, $5
        FROM companies c, platforms p
        WHERE c.id = $1 AND p.slug = $3
    `
    
    _, err := h.DB.Exec(query, dest.CompanyID, dest.Name, dest.Type, configJSON, dest.CreatedBy)
    if err != nil {
        http.Error(w, "Failed to save destination", http.StatusInternalServerError)
        return
    }
    
    // Redirect to list
    http.Redirect(w, r, "/destinations", http.StatusSeeOther)
}

func (h *MetaHandler) parseMetaForm(r *http.Request) destinations.MetaDestination {
    r.ParseForm()
    
    // Parse regions from checkboxes
    regions := strings.Join(r.Form["regions"], ",")
    
    // Parse floats with error handling
    cpmValue, _ := strconv.ParseFloat(r.FormValue("cpm_value"), 64)
    budgetDaily, _ := strconv.ParseFloat(r.FormValue("budget_daily"), 64)
    budgetTotal, _ := strconv.ParseFloat(r.FormValue("budget_total"), 64)
    
    return destinations.MetaDestination{
        BaseDestination: destinations.BaseDestination{
            Name: r.FormValue("name"),
            Type: "meta",
        },
        AccountID:    r.FormValue("account_id"),
        CampaignName: r.FormValue("campaign_name"),
        CPMValue:     cpmValue,
        BudgetDaily:  budgetDaily,
        BudgetTotal:  budgetTotal,
        Regions:      regions,
        StartDate:    r.FormValue("start_date"),
        EndDate:      r.FormValue("end_date"),
        AudienceType: r.FormValue("audience_type"),
        OptimizeFor:  r.FormValue("optimize_for"),
    }
}
