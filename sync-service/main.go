package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Hoglandets-IT/smbrsync-4-go/smbrsync"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hirochachacha/go-smb2"
)

type SyncConn struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
	Domain   string `json:"domain"`
	Share    string `json:"share"`
	Path     string `json:"path"`
}

type SyncClaim struct {
	Src     SyncConn `json:"src"`
	Dst     SyncConn `json:"dst"`
	Exclude []string `json:"exclude"`
	jwt.StandardClaims
}

func getClaims(tokenString string, secret string) (*SyncClaim, error) {

	token, err := jwt.ParseWithClaims(tokenString, &SyncClaim{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if claims, ok := token.Claims.(*SyncClaim); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err

	}
}

func sync(claims *SyncClaim) (*smbrsync.SmbRsyncResult, error) {

	srcConn, err := net.Dial("tcp", claims.Src.Server+":445")
	if err != nil {
		return nil, err
	}
	defer srcConn.Close()

	srcCredentials := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     claims.Src.Username,
			Password: claims.Src.Password,
			Domain:   claims.Src.Domain,
		},
	}

	srcDial, err := srcCredentials.Dial(srcConn)
	if err != nil {
		return nil, err
	}
	defer srcDial.Logoff()

	srcShare, err := srcDial.Mount(claims.Src.Share)
	if err != nil {
		return nil, err
	}
	defer srcShare.Umount()

	dstConn, err := net.Dial("tcp", claims.Dst.Server+":445")
	if err != nil {
		return nil, err
	}
	defer dstConn.Close()

	dstCredentials := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     claims.Dst.Username,
			Password: claims.Dst.Password,
			Domain:   claims.Dst.Domain,
		},
	}

	dstDial, err := dstCredentials.Dial(dstConn)
	if err != nil {
		return nil, err
	}
	defer dstDial.Logoff()

	dstShare, err := dstDial.Mount(claims.Dst.Share)
	if err != nil {
		return nil, err
	}
	defer dstShare.Umount()

	time.Sleep(1 * time.Second)

	sync, err := smbrsync.New(
		&smbrsync.SmbRsyncShare{
			Share:    srcShare,
			BasePath: claims.Src.Path,
		},

		&smbrsync.SmbRsyncShare{
			Share:    dstShare,
			BasePath: claims.Dst.Path,
		},

		claims.Exclude,
	)
	if err != nil {
		panic(err)
	}

	result, err := sync.Sync()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func handleSyncRequest(secret string) func(c *gin.Context) {

	return func(c *gin.Context) {

		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": err.Error(),
				"result":  nil,
			})
			return
		}

		tokenString := string(data)

		claims, err := getClaims(tokenString, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
				"result":  nil,
			})
			return
		}

		result, err := sync(claims)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": err.Error(),
				"result":  nil,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "woop woop!",
			"result": map[string]map[string]interface{}{
				"copied": {
					"number": len(result.Copied),
					"items":  result.Copied,
				},
				"skipped": {
					"number": len(result.Skipped),
					"items":  result.Skipped,
				},
				"excluded": {
					"number": len(result.Excluded),
					"items":  result.Excluded,
				},
				"mismatch": {
					"number": len(result.Mismatch),
					"items":  result.Mismatch,
				},
				"deleted": {
					"number": len(result.Deleted),
					"items":  result.Deleted,
				},
				"total": {
					"number": len(result.Copied) + len(result.Skipped) + len(result.Excluded) + len(result.Mismatch) + len(result.Deleted),
				},
			},
		})
	}
}

func main() {

	secret := os.Getenv("SYNC_SECRET")
	if secret == "" {
		panic("unable to start service, missing env \"SYNC_SECRET\".")
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.POST("/sync", handleSyncRequest(secret))

	r.Run()
}
