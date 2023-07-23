package models

import (
	"database/sql"
	"fmt"
	"time"

	"mvc/utils"

	_ "github.com/go-sql-driver/mysql"
)

type POI struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Address   string     `json:"address"`
	PoiType   string     `json:"poi_type"`
	Latitude  float64    `json:"latitude"`
	Longitude float64    `json:"longitude"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func GetPOIs() ([]POI, error) {
	db := utils.GetDB()
	query := "SELECT * FROM pois"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pois := make([]POI, 0)
	for rows.Next() {
		var poi POI
		err := rows.Scan(&poi.ID, &poi.Name, &poi.Address, &poi.PoiType, &poi.Latitude, &poi.Longitude, &poi.CreatedAt, &poi.UpdatedAt)
		if err != nil {
			return nil, err
		}
		pois = append(pois, poi)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pois, nil
}

func CreatePOI(poi POI) (POI, error) {
	db := utils.GetDB()
	timestamp := time.Now()

	query := "INSERT INTO pois (name, address, poi_type, latitude, longitude, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(query, poi.Name, poi.Address, poi.PoiType, poi.Latitude, poi.Longitude, timestamp, timestamp)
	if err != nil {
		return POI{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return POI{}, err
	}

	poi.ID = uint(id)
	poi.CreatedAt = timestamp
	poi.UpdatedAt = timestamp

	return poi, nil
}

func GetPOIByID(id uint) (POI, error) {
	db := utils.GetDB()
	query := "SELECT * FROM pois WHERE id = ?"

	poi := POI{}
	err := db.QueryRow(query, id).Scan(&poi.ID, &poi.Name, &poi.Address, &poi.PoiType, &poi.Latitude, &poi.Longitude, &poi.CreatedAt, &poi.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return POI{}, nil
		}
		return POI{}, err
	}

	return poi, nil
}

func UpdatePOI(poi POI) (POI, error) {
	db := utils.GetDB()
	timestamp := time.Now()

	query := "UPDATE pois SET name = ?, address = ?, poi_type = ?, latitude = ?, longitude = ?, updated_at = ? WHERE id = ?"
	_, err := db.Exec(query, poi.Name, poi.Address, poi.PoiType, poi.Latitude, poi.Longitude, timestamp, poi.ID)
	if err != nil {
		return POI{}, err
	}

	poi.UpdatedAt = timestamp

	return poi, nil
}

func DeletePOI(id uint) error {
	db := utils.GetDB()
	query := "DELETE FROM pois WHERE id = ?"

	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}
