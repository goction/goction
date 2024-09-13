package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"goction/internal/config"
)

// ExportGoction exports a goction to a zip file
func ExportGoction(args []string, cfg *config.Config) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: goction export <goction-name>")
	}
	name := args[0]

	srcPath := filepath.Join(cfg.GoctionsDir, name)
	destPath := fmt.Sprintf("%s.zip", name)

	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("goction '%s' does not exist", name)
	}

	zipFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		if relPath == "." {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = relPath
		if info.IsDir() {
			header.Name += "/"
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to export goction: %w", err)
	}

	fmt.Printf("Goction '%s' exported to %s\n", name, destPath)
	return nil
}

// ImportGoction imports a goction from a zip file
func ImportGoction(args []string, cfg *config.Config) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: goction import <zip-file>")
	}
	zipPath := args[0]

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	goctionName := strings.TrimSuffix(filepath.Base(zipPath), ".zip")
	destPath := filepath.Join(cfg.GoctionsDir, goctionName)

	if _, err := os.Stat(destPath); !os.IsNotExist(err) {
		return fmt.Errorf("goction '%s' already exists", goctionName)
	}

	for _, file := range reader.File {
		filePath := filepath.Join(destPath, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, file.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		srcFile, err := file.Open()
		if err != nil {
			dstFile.Close()
			return fmt.Errorf("failed to open file in zip: %w", err)
		}

		_, err = io.Copy(dstFile, srcFile)
		dstFile.Close()
		srcFile.Close()

		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	fmt.Printf("Goction '%s' imported successfully\n", goctionName)
	return nil
}
