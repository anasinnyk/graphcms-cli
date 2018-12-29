package main

import (
  "io"
  "os"
  "fmt"
  "bytes"
  "strconv"
  "net/http"
  "encoding/csv"
  "encoding/json"
  "text/template"

  "github.com/spf13/cobra"
)

var token string
var url string

func main() {
  var rootCmd = &cobra.Command{
    Use: "graphcms",
    Short: "GraphCMS",
    Long: `No tech know-how needed! GraphCMS gives content creators the tools to easily create content of any shape.
Make use of role based publishing workflows or translate your content to any language you like.`,
  }


  importCmd.Flags().StringVarP(&url, "url", "u", "", "GraphCMS URL")
  importCmd.Flags().StringVarP(&token, "token", "t", "", "GraphCMS Token")

  rootCmd.AddCommand(importCmd)
  rootCmd.Execute()
}

var importCmd = &cobra.Command{
  Use:   "import [model] [path to csv]",
  Short: "Import model data to GraphCMS",
  Args: cobra.MinimumNArgs(2),
  Run: func(cmd *cobra.Command, args []string) {
    if url == "" || token == "" {
      fmt.Println("Options Url and Token are required")
  		os.Exit(1)
    }
    model := args[0]
    path := args[1]


    file, err := os.Open(path)
  	if err != nil {
  		fmt.Println("Cannot open csv file:", err)
  		os.Exit(1)
  	}
  	defer file.Close()
  	reader := csv.NewReader(file)
    var header []string
    var records []map[string]interface{}
  	lineCount := 0
  	for {
  		record, err := reader.Read()
  		if err == io.EOF {
  			break
  		} else if err != nil {
  			fmt.Println("File Read Error:", err)
  			os.Exit(1)
  		}

      if lineCount == 0 {
        header = record
      } else {
        m := make(map[string]interface{})
    		for i := 0; i < len(record); i++ {
          m[header[i]], err = strconv.ParseFloat(record[i], 64)
          if err != nil {
            m[header[i]], err = strconv.ParseInt(record[i], 10, 64)
            if err != nil {
              m[header[i]], err = strconv.ParseBool(record[i])
              if err != nil {
                m[header[i]] = strconv.Quote(record[i])
              }
            }
          }
    		}
        records = append(records, m)
      }
  		lineCount += 1
  	}

    const requestTemplate = `
      mutation {
        create{{.Name}}(data: {
            {{ range $k, $v := .Values -}}
              {{ printf "%s: %s,\n" $k $v }}
            {{- end }}
          }) {
          id
        }
      }
    `

  	type Model struct {
  		Name string
      Values map[string]interface{}
  	}

  	type Query struct {
      Query string `json:"query"`
    }
    client := &http.Client{}

    for _, record := range records {
      data := Model{
        Name: model,
        Values: record,
    	}
      requestBody := new(bytes.Buffer)
      t := template.Must(template.New("request").Parse(requestTemplate))
      if err := t.Execute(requestBody, data); err != nil {
        fmt.Println("execute template:", err)
        os.Exit(1)
      }
      q := &Query{
        Query: requestBody.String(),
      }
      d, _ := json.Marshal(q)
      req, _ := http.NewRequest("POST", url, bytes.NewBuffer(d))
      req.Header.Set("Content-Type", "application/json")
      req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
      resp, err := client.Do(req)
      if err != nil {
        fmt.Print("E")
      }
      if resp.StatusCode == http.StatusOK {
        fmt.Print(".")
      } else {
        fmt.Print("F")
      }
    }
  },
}
