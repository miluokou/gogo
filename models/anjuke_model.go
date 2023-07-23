package models

import (
// 	"database/sql"
// 	"fmt"
	"mvc/utils"
	_ "github.com/go-sql-driver/mysql"
)

type PropertyData struct {
	ID            uint   `json:"id"`
	YearInfo      string `json:"year_info"`
	CommunityName string `json:"community_name"`
	AddressInfo   string `json:"address_info"`
	PricePerSqm   string `json:"price_per_sqm"`
	PageNumber    string `json:"page_number"`
	Deal          string `json:"deal"`
	City          string `json:"city"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func GetOriginData() ([]PropertyData, error) {
	db := utils.GetDB2()
	query := "SELECT * FROM djangoproject_propertydata WHERE deal = '0' LIMIT 100"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	properties := make([]PropertyData, 0)
	for rows.Next() {
		var prop PropertyData
		err := rows.Scan(&prop.ID,
			&prop.YearInfo,
			&prop.CommunityName,
			&prop.AddressInfo,
			&prop.PricePerSqm,
			&prop.PageNumber,
			&prop.Deal,
			&prop.City,
			&prop.CreatedAt,
			&prop.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		properties = append(properties, prop)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return properties, nil
}