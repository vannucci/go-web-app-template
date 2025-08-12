package destinations

import (
    "encoding/json"
    "fmt"
    "strings"
)

type MetaDestination struct {
    BaseDestination
    
    // Meta-specific fields
    AccountID      string  `json:"account_id"`
    CampaignName   string  `json:"campaign_name"`
    CPMValue       float64 `json:"cpm_value"`
    BudgetDaily    float64 `json:"budget_daily"`
    BudgetTotal    float64 `json:"budget_total"`
    Regions        string  `json:"regions"`
    StartDate      string  `json:"start_date"`
    EndDate        string  `json:"end_date"`
    AudienceType   string  `json:"audience_type"`
    OptimizeFor    string  `json:"optimize_for"`
}

func (m *MetaDestination) Validate() error {
    var errors []string
    
    // Required fields
    if m.AccountID == "" {
        errors = append(errors, "Account ID is required")
    }
    if !strings.HasPrefix(m.AccountID, "act_") {
        errors = append(errors, "Account ID must start with 'act_'")
    }
    if m.CampaignName == "" {
        errors = append(errors, "Campaign name is required")
    }
    
    // Value validations
    if m.CPMValue <= 0 {
        errors = append(errors, "CPM value must be positive")
    }
    if m.CPMValue > 1000 {
        errors = append(errors, "CPM value seems too high (max: $1000)")
    }
    
    if m.BudgetDaily <= 0 {
        errors = append(errors, "Daily budget must be positive")
    }
    if m.BudgetTotal < m.BudgetDaily {
        errors = append(errors, "Total budget must be >= daily budget")
    }
    
    // Date validation
    if m.StartDate == "" {
        errors = append(errors, "Start date is required")
    }
    if m.EndDate != "" && m.EndDate < m.StartDate {
        errors = append(errors, "End date must be after start date")
    }
    
    // Regions validation
    if m.Regions == "" {
        errors = append(errors, "At least one region must be selected")
    }
    validRegions := map[string]bool{
        "US": true, "CA": true, "UK": true, "FR": true, "DE": true,
    }
    for _, region := range strings.Split(m.Regions, ",") {
        region = strings.TrimSpace(region)
        if region != "" && !validRegions[region] {
            errors = append(errors, fmt.Sprintf("Invalid region: %s", region))
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf(strings.Join(errors, "; "))
    }
    return nil
}

func (m *MetaDestination) GetType() string {
    return "meta"
}

func (m *MetaDestination) GetDisplayName() string {
    return "Meta (Facebook/Instagram)"
}

func (m *MetaDestination) ToJSON() ([]byte, error) {
    return json.Marshal(m)
}

func (m *MetaDestination) FromJSON(data []byte) error {
    return json.Unmarshal(data, m)
}
