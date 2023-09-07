package models

import (
	_ "github.com/go-sql-driver/mysql"
	"mvc/service"
	"mvc/utils"
)

type PropertyData struct {
	ID            uint     `json:"id"`
	YearInfo      *string  `json:"year_info"`
	CommunityName *string  `json:"community_name"`
	AddressInfo   *string  `json:"address_info"`
	PricePerSqm   *float64 `json:"price_per_sqm"`
	PageNumber    *int     `json:"page_number"`
	HouseHolds    *string  `json:"households"`
	Deal          *string  `json:"deal"`
	City          *string  `json:"city"`
	QuText        *string  `json:"qu_text"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

func GetOriginData() ([]PropertyData, error) {
	db := utils.GetDB2()
	query := "SELECT id,year_info,community_name,address_info,price_per_sqm,page_number,households,deal,city,qu_text,created_at,updated_at FROM djangoproject_propertydata WHERE deal = '0' LIMIT 2"

	rows, err := db.Query(query)
	if err != nil {
		service.LogInfo("3")
		service.LogInfo(err)
		return nil, err
	}
	defer rows.Close()

	properties := make([]PropertyData, 0)
	for rows.Next() {
		var prop PropertyData
		err := rows.Scan(
			&prop.ID,
			&prop.YearInfo,
			&prop.CommunityName,
			&prop.AddressInfo,
			&prop.PricePerSqm,
			&prop.PageNumber,
			&prop.HouseHolds,
			&prop.Deal,
			&prop.City,
			&prop.QuText,
			&prop.CreatedAt,
			&prop.UpdatedAt,
		)
		if err != nil {
			service.LogInfo(err)
			return nil, err
		}

		// 处理可能为 NULL 的字段
		if prop.YearInfo == nil {
			empty := ""
			prop.YearInfo = &empty
		}
		if prop.CommunityName == nil {
			empty := ""
			prop.CommunityName = &empty
		}
		if prop.AddressInfo == nil {
			empty := ""
			prop.AddressInfo = &empty
		}
		if prop.PricePerSqm == nil {
			zero := 0.0
			prop.PricePerSqm = &zero
		}
		if prop.PageNumber == nil {
			zero := 0
			prop.PageNumber = &zero
		}
		if prop.HouseHolds == nil {
			empty := ""
			prop.HouseHolds = &empty
		}
		if prop.Deal == nil {
			empty := ""
			prop.Deal = &empty
		}
		if prop.City == nil {
			empty := ""
			prop.City = &empty
		}
		if prop.QuText == nil {
			empty := ""
			prop.QuText = &empty
		}

		properties = append(properties, prop)
	}

	if err := rows.Err(); err != nil {
		service.LogInfo(err)
		return nil, err
	}

	return properties, nil
}

/**
 * 通过id 去删除 已经存入数据库中的数据
 */
func (prop *PropertyData) Delete() error {
	db := utils.GetDB2()
	query := "DELETE FROM djangoproject_propertydata WHERE id = ?"

	_, err := db.Exec(query, prop.ID)
	if err != nil {
		service.LogInfo(err)
		return err
	}
	return nil
}
