// handler/util.go

package handler

import (
	"clusterviz/internal/pkg/configurations"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
)

// Cluster represents a cluster entity.
type Cluster struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// Add more fields as per your cluster data structure
}

// EndPointHandler handles endpoints logic.
type EndPointHandler struct {
	db *sql.DB
}

// NewEndPointHandler creates a new instance of EndPointHandler.
func NewEndPointHandler(conf *configurations.ServiceConfigurations) (*EndPointHandler, error) {
	db, err := sql.Open("mysql", conf.DBConnectionString)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &EndPointHandler{db: db}, nil
}

// GetClusters fetches all clusters from the database.
func (h *EndPointHandler) GetClusters(c *gin.Context) {
    clusters, err := h.fetchClustersFromDB()
    if err != nil {
        // Handle error and return appropriate response to the client
        // ...
        return
    }

    // Return clusters to the client
    c.JSON(http.StatusOK, clusters)
}
// GetClusterByID fetches a specific cluster by ID from the database.
func (h *EndPointHandler) GetClustersId(c *gin.Context, id int) {
    cluster, err := h.fetchClusterByIDFromDB(id)
    if err != nil {
        // Handle error and return appropriate response to the client
        // ...
        return
    }

    // Return cluster to the client
    c.JSON(http.StatusOK, cluster)
}
func (h *EndPointHandler) fetchClustersFromDB() ([]Cluster, error) {
    rows, err := h.db.Query("SELECT id, name FROM clusters")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var clusters []Cluster
    for rows.Next() {
        var cluster Cluster
        if err := rows.Scan(&cluster.ID, &cluster.Name); err != nil {
            return nil, err
        }
        clusters = append(clusters, cluster)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return clusters, nil
}
func (h *EndPointHandler) fetchClusterByIDFromDB(id int) (Cluster, error) {
    var cluster Cluster
    err := h.db.QueryRow("SELECT id, name FROM clusters WHERE id = ?", id).Scan(&cluster.ID, &cluster.Name)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return Cluster{}, fmt.Errorf("cluster with ID %d not found", id)
        }
        return Cluster{}, err
    }
    return cluster, nil
}

// Authenticator is a middleware function for JWT-based authentication.
func Authenticator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the request header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token not provided",
			})
			c.Abort()
			return
		}

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify the token signing method and return the secret key
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Invalid token signing method")
			}
			return []byte("your_secret_key"), nil // Replace "your_secret_key" with your actual secret key
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}


		c.Next()
	}
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