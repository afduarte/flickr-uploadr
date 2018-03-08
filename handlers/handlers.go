package handlers

import (
	"os"
	"io"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/afduarte/flickr-uploadr/DB"
)

func Upload(c echo.Context) error {
	// Read form fields
	title := c.FormValue("title")
	description := c.FormValue("description")
	tags := c.FormValue("tags")
	checksum := c.FormValue("checksum")

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
	if err = DB.AddPhoto(filename.String(), checksum); err != nil {
		return err
	}
	return c.HTML(
		http.StatusOK,
		fmt.Sprintf("<p>File %s uploaded successfully with fields title=%s, desc=%s and tags:=%s.</p>",
			file.Filename,
			title,
			description,
			tags,
		),
	)
}

func CheckPhotoExists(c echo.Context) error {
	checksum := c.QueryParam("checksum")
	if DB.CheckPhotoExists(checksum) {
		return c.String(http.StatusNoContent, "")
	} else {
		return c.String(http.StatusConflict, "")
	}
}
