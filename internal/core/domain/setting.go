package domain

// Setting เก็บค่าตั้งค่าต่างๆ
type Setting struct {
	SettingName  string `gorm:"primaryKey" json:"setting_name"`
	SettingValue string `json:"setting_value"`
	
	// Optional fields for easy UI generation
	Group       string `json:"group"`       // e.g., "general", "theme", "booking"
	Type        string `json:"type"`        // e.g., "text", "number", "boolean", "image", "color"
	Label       string `json:"label"`       // e.g., "System Name"
	Description string `json:"description"` // e.g., "The name displayed on the login screen"
}