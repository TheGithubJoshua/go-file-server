package main

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/list", listFilesHandler)
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./files"))))

	port := ":8080"
	println("File server running on http://localhost" + port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}
}

const filesTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>List of Files</title>
</head>
<body>
    <h1>List of Files in the Folder</h1>
    <ul>
        {{range .Files}}
        <li>{{.}}</li>
        {{end}}
    </ul>
</body>
<style>
h1 {
    text-align: center;
    margin: 30px 0;
    color: #ff8800;
}
</style>
</html>`

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/upload.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error reading form file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Ensure the "files" directory exists
	if _, err := os.Stat("files"); os.IsNotExist(err) {
		err := os.Mkdir("files", 0755) // Creates the "files" directory with read/write permissions for the owner
		if err != nil {
			http.Error(w, "Error creating the 'files' directory", http.StatusInternalServerError)
			return
		}
	}

	// Save the uploaded file to the "files" directory
	// You can use a unique filename to prevent overwriting existing files
	// For simplicity, we'll use the original filename
	filePath := "./files/" + handler.Filename
	err = saveUploadedFile(file, filePath)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}

	// Redirect back to the upload form
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func saveUploadedFile(fileContent io.Reader, destination string) error {
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, fileContent)
	return err
}

func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	// Replace "/path/to/folder" with the actual path of the folder you want to list files from.
	folderPath := "./files"

	// Get the list of files in the folder
	files, err := getFilesInFolder(folderPath)
	if err != nil {
		http.Error(w, "Error listing files", http.StatusInternalServerError)
		return
	}

	// Parse the template
	tmpl, err := template.New("files_template").Parse(filesTemplate)
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	// Execute the template with the list of files and write the response
	err = tmpl.Execute(w, struct{ Files []string }{Files: files})
	if err != nil {
		http.Error(w, "Error generating HTML", http.StatusInternalServerError)
		return
	}
}

// Function to get the list of files in the folder
func getFilesInFolder(folderPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

