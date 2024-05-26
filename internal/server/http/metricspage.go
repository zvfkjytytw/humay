package humayhttpserver

import (
	"fmt"
	"html/template"
	"net/http"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	metricsHTML = `
    <!DOCTYPE HTML>
    <html>
        <head>
            <meta charset="utf-8">
            <title>HUMAY</title> 
        </head>
        <body>
            <h1>METRICS</h1>
            <div class="metrics">
                {{range .}}
                <div>
                    <h3>{{.Type}}</h3>
                    <table>
                        {{range .Metrics}}
                        <tr>
                            <td>{{.Name}}</td>
                            <td>{{.Value}}</td> 
                        </tr>
                        {{else}}
                        <tr><td>no metrics</td></tr>
                        {{end}}
                    </table>
                </div>
                {{end}}
            </div>
        </body>
        <style type="text/css">
            div.metrics {
                display: flex;
            }
            div.metrics div {
                padding: 0 20px;
            }
            table {
                width: 100%;
                border: 1px solid #dddddd;
                border-collapse: collapse; 
            }
            table tr {
                font-weight: bold;
                background: #efefef;
            }
            table td {
                border: 1px solid #dddddd;
                padding: 5px;
            }
        </style>
    </html>
    `
)

type Metric struct {
	Name  string
	Value string
}

type Monitoring struct {
	Type    string
	Metrics []Metric
}

func (h *HTTPServer) metricsPage(w http.ResponseWriter, r *http.Request) {
	allMetrics := h.storage.GetAllMetrics()
	data := make([]Monitoring, 0, len(allMetrics))
	caser := cases.Title(language.English)

	for mType, metrics := range allMetrics {
		mList := make([]Metric, 0, len(metrics))
		for name, value := range metrics {
			mList = append(
				mList,
				Metric{
					Name:  name,
					Value: value,
				},
			)
		}
		data = append(
			data,
			Monitoring{
				Type:    caser.String(mType),
				Metrics: mList,
			},
		)
	}

	htmlTemplate, err := template.New("metricsPage").Parse(metricsHTML)
	if err != nil {
		fmt.Printf("Parsing error: %w\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed parsing template"))
		return
	}

	err = htmlTemplate.Execute(w, data)
	if err != nil {
		fmt.Printf("Load error: %w\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed load metrics page"))
	}
}
