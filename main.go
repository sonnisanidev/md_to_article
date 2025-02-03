package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type HeadlineInfo struct {
	Level   int
	Content string
}

func replaceContentInTemplate(mdFilePath, htmlTemplatePath, outputFolderPath string) error {
	headlines, text, image, err := extractContent(mdFilePath)
	if err != nil {
		return fmt.Errorf("error extracting content: %v", err)
	}
	fmt.Println("Extracted headlines:", headlines)
	fmt.Println("Extracted text:", text)
	fmt.Println("Extracted image:", image)

	templateContent, err := os.ReadFile(htmlTemplatePath)
	if err != nil {
		return fmt.Errorf("error reading HTML template: %v", err)
	}
	fmt.Println("HTML template content:", string(templateContent))

	baseName := filepath.Base(mdFilePath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	newContent := replaceHeadlines(string(templateContent), headlines)
	newContent = replaceText(newContent, text)
	newContent = replaceImage(newContent, image)
	newContent = strings.Replace(newContent, "Placeholder Headline", baseName, 1) // Replace the first occurrence with the file name
	newContent = strings.Replace(newContent, "Placeholder Title", baseName, 1)    // Replace the title with the file name
	fmt.Println("Content after replacing:", newContent)

	err = os.MkdirAll(outputFolderPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating output folder: %v", err)
	}

	outputPath := filepath.Join(outputFolderPath, baseName+".html")
	err = os.WriteFile(outputPath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing output file: %v", err)
	}

	fmt.Printf("HTML file with updated content created at: %s\n", outputPath)
	return nil
}

func extractContent(filePath string) ([]HeadlineInfo, string, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", "", fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var headlines []HeadlineInfo
	var text strings.Builder
	var image string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			level := strings.Count(line, "#")
			headline := strings.TrimSpace(strings.TrimLeft(line, "# "))
			headlines = append(headlines, HeadlineInfo{Level: level, Content: headline})
		} else if strings.HasPrefix(line, "![") {
			image = strings.TrimSpace(strings.TrimPrefix(line, "![Picture]"))
			image = strings.Trim(image, "()")
		} else {
			text.WriteString(line + " ")
		}
	}
	return headlines, strings.TrimSpace(text.String()), image, nil
}

func replaceHeadlines(content string, headlines []HeadlineInfo) string {
	for _, headline := range headlines {
		placeholder := "Placeholder Headline"
		replacement := fmt.Sprintf("<h%d>%s</h%d>", headline.Level, headline.Content, headline.Level)
		content = strings.Replace(content, placeholder, replacement, 1)
	}
	return content
}

func replaceText(content, text string) string {
	return strings.Replace(content, "Test text", "<p>"+text+"</p>", 1)
}

func replaceImage(content, imagePath string) string {
	return strings.Replace(content, "Test text", fmt.Sprintf("<img src=\"%s\" alt=\"Picture\">", imagePath), 1)
}

// Print dates
func printFileCreationDate(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("error getting file info: %v", err)
	}

	creationTime := fileInfo.ModTime()
	fmt.Printf("File creation date: %s\n", creationTime.Format("Monday, January 02, 2006, 3:04 PM MST"))

	return nil
}

// Filtering
func organizeFilesByHour(sourceFolder, destinationFolder string) error {
	return filepath.Walk(sourceFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			creationTime := info.ModTime()
			yearFolder := filepath.Join(destinationFolder, creationTime.Format("2006"))
			monthFolder := filepath.Join(yearFolder, creationTime.Format("01-January"))

			err := os.MkdirAll(monthFolder, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating folder: %v", err)
			}

			// Add hour to filename
			hourPrefix := creationTime.Format("15") // 24-hour format
			newFileName := hourPrefix + "_" + info.Name()
			destPath := filepath.Join(monthFolder, newFileName)

			sourceFile, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("error opening source file: %v", err)
			}
			defer sourceFile.Close()

			destFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("error creating destination file: %v", err)
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, sourceFile)
			if err != nil {
				return fmt.Errorf("error copying file: %v", err)
			}

			fmt.Printf("Moved %s to %s\n", path, destPath)
		}
		return nil
	})
}

func main() {
	mdFolderPath := "contents/mdfiles"
	htmlTemplatePath := "contents/layout/article_layout.html"
	outputFolderPath := "contents/articles"

	err := filepath.Walk(mdFolderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			err := replaceContentInTemplate(path, htmlTemplatePath, outputFolderPath)
			if err != nil {
				fmt.Printf("Error processing %s: %v\n", path, err)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through directory: %v\n", err)
	}

	errdate := printFileCreationDate("contents/articles/blog1.html")
	if errdate != nil {
		fmt.Printf("Error: %v\n", errdate)
	}

	sourceFolderFilter := "contents/articles"
	destinationFolderFilter := "../Homepage2/homepage/output"

	errFilter := organizeFilesByHour(sourceFolderFilter, destinationFolderFilter)
	if errFilter != nil {
		fmt.Printf("Error organizing files: %v\n", errFilter)
	}

}
