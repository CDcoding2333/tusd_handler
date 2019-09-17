package main

import (
	"CDcoding2333/tusd_handler/cloud"
	mclient "CDcoding2333/tusd_handler/cloud/minio-client"
	"CDcoding2333/tusd_handler/local"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go"
)

var s3Client *s3.S3
var minioClient *minio.Client
var bucket string

var lfs *local.FileSystem
var localfileServer http.Handler

// main is only for test
// use cloud store or local store by imports cloud or local package
func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(gin.Recovery())
	router.GET("/ping", ping)

	conf := &mclient.Config{
		MinioEndpoint:  "127.0.0.1:9000",
		MinioRegion:    "us-east-1",
		MinioAppID:     "AKIAIOSFODNN7EXAMPLE",
		MinioAppSecret: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		MinioToken:     "",
	}

	s3Client = mclient.NewS3Client(conf)
	minioClient = mclient.NewMinioClient(conf)
	bucket = "bucket1"
	exist, err := minioClient.BucketExists(bucket)
	if err != nil {
		log.Fatalf("check bucket exist error:%v", err)
	}

	if !exist {
		if err := minioClient.MakeBucket(bucket, conf.MinioRegion); err != nil {
			log.Fatalf("minio create openbucket error:%v", err)
		}
	}

	lfs = local.NewFileSystem("./data/bucket1", false)
	localfileServer = http.FileServer(lfs)

	v1 := router.Group("/tusd/v1")
	{
		cloud := v1.Group("/cloud")
		{
			cloud.Any("/upload/*res", handles3tusd())
			cloud.GET("/static/:id", gets3)
			cloud.GET("/static/:id/download", downloads3)
		}

		local := v1.Group("/local")
		{
			local.Any("/upload/*res", handlelocaltusd())
			local.GET("/static/:id", getlocal)
			local.GET("/static/:id/download", downloadlocal)
		}
	}

	go router.Run(fmt.Sprintf(":%d", 9999))
	log.Printf("server start %d", 9999)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", 9999),
		Handler: router,
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(
		ch,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	<-ch
	if err := server.Close(); err != nil {
		log.Fatal("Server Close:", err)
	}
}

func handles3tusd() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		conf := &cloud.Config{
			Router:  "/tusd/v1/cloud/upload/",
			Bucket:  bucket,
			Service: s3Client,
		}

		tusdHandler, err := cloud.NewHandler(conf)
		if err != nil {
			log.Panic(err)
			return
		}
		http.StripPrefix("/tusd/v1/cloud/upload/", tusdHandler).ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func handlelocaltusd() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		conf := &local.Config{
			Router:   "/tusd/v1/local/upload/",
			FilePath: "./data/bucket1",
		}

		tusdHandler, err := local.NewHandler(conf)
		if err != nil {
			log.Panic(err)
			return
		}
		http.StripPrefix("/tusd/v1/local/upload/", tusdHandler).ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func gets3(ctx *gin.Context) {

	key := ctx.Param("id")

	object, err := minioClient.GetObject(bucket, key, minio.GetObjectOptions{})
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	defer object.Close()

	minioInfo, err := object.Stat()
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	if contentType := minioInfo.Metadata.Get("X-Amz-Meta-Type"); contentType != "" {
		ctx.Writer.Header().Set("Content-Type", contentType)
	}
	SetExpires(ctx)
	http.ServeContent(ctx.Writer, ctx.Request, "object", minioInfo.LastModified, object)
}

func downloads3(ctx *gin.Context) {
	key := ctx.Param("id")

	object, err := minioClient.GetObject(bucket, key, minio.GetObjectOptions{})
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	defer object.Close()

	minioInfo, err := object.Stat()
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	name := key
	if contentName := minioInfo.Metadata.Get("X-Amz-Meta-Name"); contentName != "" {
		name = contentName
	}

	ctx.Writer.Header().Set("Content-Type", "application/octet-stream")
	ctx.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s;filename*=utf8''%s", url.QueryEscape(name), url.QueryEscape(name)))
	http.ServeContent(ctx.Writer, ctx.Request, "object", minioInfo.LastModified, object)
}

func getlocal(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.Request.URL.Path = id

	fileInfo, err := lfs.FileInfo(id)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	contentType := "application/octet-stream"
	if cType, ok := fileInfo.MetaData["type"]; ok {
		contentType = cType
	}
	SetExpires(ctx)
	ctx.Writer.Header().Set("Content-Type", contentType)
	localfileServer.ServeHTTP(ctx.Writer, ctx.Request)
}

func downloadlocal(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.Request.URL.Path = id

	fileInfo, err := lfs.FileInfo(id)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	fileNmae := id
	if name, ok := fileInfo.MetaData["name"]; ok {
		fileNmae = name
	}

	ctx.Writer.Header().Set("Content-Type", "application/octet-stream")
	ctx.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s;filename*=utf8''%s", url.QueryEscape(fileNmae), url.QueryEscape(fileNmae)))
	localfileServer.ServeHTTP(ctx.Writer, ctx.Request)
}

func ping(c *gin.Context) {
	c.Status(http.StatusOK)
}

//SetExpires ...
func SetExpires(ctx *gin.Context) {
	cacheUntil := time.Now().AddDate(0, 0, 1).Format(http.TimeFormat)
	ctx.Writer.Header().Set("Cache-Control", "max-age:604800, public")
	ctx.Writer.Header().Set("Expires", cacheUntil)
}
