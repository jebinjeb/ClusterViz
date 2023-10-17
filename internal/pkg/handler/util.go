package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Cluster struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type VisualizationData struct {
	ID   int    `json:"id"`
	Data string `json:"data"`
}

func GetClustersFromDB(ctx context.Context, db *sql.DB) ([]Cluster, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, name FROM clusters")
	if err != nil {
		return nil, fmt.Errorf("error querying clusters: %v", err)
	}
	defer rows.Close()

	var clusters []Cluster
	for rows.Next() {
		var cluster Cluster
		if err := rows.Scan(&cluster.ID, &cluster.Name); err != nil {
			return nil, fmt.Errorf("error scanning clusters: %v", err)
		}
		clusters = append(clusters, cluster)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clusters: %v", err)
	}
	return clusters, nil
}

func GetClusterByIDFromDB(ctx context.Context, db *sql.DB, id int) (Cluster, error) {
	var cluster Cluster
	err := db.QueryRowContext(ctx, "SELECT id, name FROM clusters WHERE id = ?", id).Scan(&cluster.ID, &cluster.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Cluster{}, fmt.Errorf("cluster with ID %d not found", id)
		}
		return Cluster{}, fmt.Errorf("error scanning cluster: %v", err)
	}
	return cluster, nil
}

func GetClusterVizDataFromDB(ctx context.Context, db *sql.DB) ([]VisualizationData, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, data FROM visualization_data")
	if err != nil {
		return nil, fmt.Errorf("error querying visualization data: %v", err)
	}
	defer rows.Close()

	var vizData []VisualizationData
	for rows.Next() {
		var data VisualizationData
		if err := rows.Scan(&data.ID, &data.Data); err != nil {
			return nil, fmt.Errorf("error scanning visualization data: %v", err)
		}
		vizData = append(vizData, data)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating visualization data: %v", err)
	}
	return vizData, nil
}




/*package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"clusterviz/api"
	gerrors "clusterviz/internal/pkg/gerror"
	"net/http"
)

func respondWithError(ginCtx *gin.Context, err error, msg string) {
	var statusCode int

	defaultMsg := "Internal server error, please check with Marketplace Admin"

	switch gerrors.GetErrorType(err) { // nolint:exhaustive
	case gerrors.ValidationFailed, gerrors.BadRequest:
		log.Errorf("error while processing request: %v msg :%s", err, msg)

		statusCode = http.StatusBadRequest
		defaultMsg = "Re-verify the provided request"
	case gerrors.NotFound:
		log.Errorf("EmailAccount not present: error : %v msg :%s", err, msg)

		statusCode = http.StatusNotFound
		msg = "Requested resource not found"

	case gerrors.TokenNotFound, gerrors.AuthenticationFailed:
		log.Errorf("Connection issue : error : %v msg :%s", err, msg)

		statusCode = http.StatusUnauthorized
		msg = "Invalid token/Session, please login"
	default:
		log.Errorf("Internal issue while processing request : %v , msg : %s", err, msg)

		statusCode = http.StatusInternalServerError
	}

	if msg == "" {
		msg = defaultMsg
	}

	ginCtx.JSON(statusCode, api.ErrorModel{
		Message:   msg,
		ErrorCode: fmt.Sprint(statusCode),
	})
}
*/