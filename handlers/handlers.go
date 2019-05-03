package handlers

import (
	"os"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/afduarte/flickr-uploadr/model"
	"image"
	"github.com/nfnt/resize"
	"image/jpeg"
	_ "image/gif"
	_ "image/png"
	_ "golang.org/x/image/tiff"
	"strconv"
)

func Upload(c echo.Context) error {
	// Read form fields
	title := c.FormValue("title")
	description := c.FormValue("description")
	tagList := c.FormValue("tags")
	checksum := c.FormValue("checksum")
	tags := strings.Split(tagList, ",")

	//-----------
	// Read file
	//-----------
	// generate a random uuid for the filename
	filename, _ := uuid.NewV4()
	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	dst, err := os.Create("files/" + filename.String())
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	// Seek dst back to 0,0
	if _, err := dst.Seek(0, 0); err != nil {
		return err
	}

	// Thumb
	thumb, err := os.Create("files/" + filename.String() + "_thumb")
	if err != nil {
		return err
	}
	defer thumb.Close()
	err = CreateThumb(dst, thumb)
	if err != nil {
		return err
	}

	photo := model.Photo{
		ID:          filename.String(),
		Checksum:    checksum,
		Title:       title,
		Description: description,
		Tags:        tags,
	}
	if err = model.AddPhoto(&photo); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, photo)
}

func CheckPhotoExists(c echo.Context) error {
	checksum := c.QueryParam("checksum")
	if model.CheckPhotoExists(checksum) {
		return c.String(http.StatusOK, "")
	} else {
		return c.String(http.StatusNotFound, "")
	}
}

func AddJob(c echo.Context) error {
	id := c.FormValue("id")
	whenStr := c.FormValue("when")
	unixTime, err := strconv.ParseInt(whenStr, 10, 64)
	if err != nil {
		return err
	}
	if err := model.AddJob(id, unixTime); err != nil {
		return err
	}
	return c.String(http.StatusOK, "")
}

// Helpers

func CreateThumb(full io.Reader, thumb io.Writer) error {
	img, _, err := image.Decode(full)
	if err != nil {
		return err
	}
	small := resize.Thumbnail(500, 500, img, resize.NearestNeighbor)
	err = jpeg.Encode(thumb, small, nil)
	if err != nil {
		return err
	}
	return nil
}
