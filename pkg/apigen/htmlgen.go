package apigen

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

// HTMLGenerator creates dynamic HTML pages from JSON data
type HTMLGenerator struct {
	Config HTMLConfig
}

// HTMLConfig holds configuration for HTML generation
type HTMLConfig struct {
	OutputDir    string
	Theme        string
	EnableSearch bool
	EnableSort   bool
	Title        string
	CDNLibs      bool
}

// DefaultHTMLConfig returns default HTML generation configuration
func DefaultHTMLConfig() HTMLConfig {
	return HTMLConfig{
		OutputDir:    "./html",
		Theme:        "default",
		EnableSearch: true,
		EnableSort:   true,
		Title:        "API Data Viewer",
		CDNLibs:      true,
	}
}

// NewHTMLGenerator creates a new HTML generator
func NewHTMLGenerator(config HTMLConfig) *HTMLGenerator {
	return &HTMLGenerator{
		Config: config,
	}
}

// GenerateInteractivePages creates interactive HTML pages from JSON data
func (hg *HTMLGenerator) GenerateInteractivePages(data interface{}) error {
	if err := os.MkdirAll(hg.Config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate main dashboard
	if err := hg.generateDashboard(data); err != nil {
		return fmt.Errorf("failed to generate dashboard: %w", err)
	}

	// Generate data table view
	if err := hg.generateTableView(data); err != nil {
		return fmt.Errorf("failed to generate table view: %w", err)
	}

	// Generate chart view
	if err := hg.generateChartView(data); err != nil {
		return fmt.Errorf("failed to generate chart view: %w", err)
	}

	// Generate raw JSON view
	if err := hg.generateJSONView(data); err != nil {
		return fmt.Errorf("failed to generate JSON view: %w", err)
	}

	// Copy static assets
	if err := hg.generateStaticAssets(); err != nil {
		return fmt.Errorf("failed to generate static assets: %w", err)
	}

	return nil
}

func (hg *HTMLGenerator) generateDashboard(data interface{}) error {
	dashboardTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Dashboard</title>
    {{if .CDNLibs}}
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script src="https://code.jquery.com/jquery-3.6.0.min.js"></script>
    {{end}}
    <style>
        .dashboard-card { margin-bottom: 20px; }
        .metric-value { font-size: 2em; font-weight: bold; color: #007bff; }
        .metric-label { color: #6c757d; }
        .nav-tabs .nav-link.active { background-color: #007bff; color: white; }
    </style>
</head>
<body>
    <div class="container-fluid">
        <nav class="navbar navbar-expand-lg navbar-dark bg-primary mb-4">
            <div class="container">
                <a class="navbar-brand" href="#">{{.Title}}</a>
                <ul class="navbar-nav">
                    <li class="nav-item"><a class="nav-link active" href="index.html">Dashboard</a></li>
                    <li class="nav-item"><a class="nav-link" href="table.html">Table View</a></li>
                    <li class="nav-item"><a class="nav-link" href="charts.html">Charts</a></li>
                    <li class="nav-item"><a class="nav-link" href="json.html">Raw JSON</a></li>
                </ul>
            </div>
        </nav>

        <div class="container">
            <div class="row">
                <div class="col-md-3">
                    <div class="card dashboard-card">
                        <div class="card-body text-center">
                            <div class="metric-value" id="record-count">{{.RecordCount}}</div>
                            <div class="metric-label">Total Records</div>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card dashboard-card">
                        <div class="card-body text-center">
                            <div class="metric-value" id="field-count">{{.FieldCount}}</div>
                            <div class="metric-label">Fields</div>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card dashboard-card">
                        <div class="card-body text-center">
                            <div class="metric-value">{{.LastUpdated}}</div>
                            <div class="metric-label">Last Updated</div>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card dashboard-card">
                        <div class="card-body text-center">
                            <div class="metric-value text-success">●</div>
                            <div class="metric-label">Status: Active</div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="row">
                <div class="col-md-8">
                    <div class="card">
                        <div class="card-header">
                            <h5>Data Preview</h5>
                        </div>
                        <div class="card-body">
                            <div id="data-preview"></div>
                        </div>
                    </div>
                </div>
                <div class="col-md-4">
                    <div class="card">
                        <div class="card-header">
                            <h5>Quick Actions</h5>
                        </div>
                        <div class="card-body">
                            <div class="d-grid gap-2">
                                <button class="btn btn-primary" onclick="refreshData()">Refresh Data</button>
                                <button class="btn btn-outline-secondary" onclick="exportData('json')">Export JSON</button>
                                <button class="btn btn-outline-secondary" onclick="exportData('csv')">Export CSV</button>
                                <a href="table.html" class="btn btn-outline-info">View Table</a>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        const rawData = {{.JSONData}};
        
        function refreshData() {
            location.reload();
        }
        
        function exportData(format) {
            const dataStr = format === 'json' ? JSON.stringify(rawData, null, 2) : convertToCSV(rawData);
            const blob = new Blob([dataStr], {type: format === 'json' ? 'application/json' : 'text/csv'});
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'data.' + format;
            a.click();
        }
        
        function convertToCSV(data) {
            if (!Array.isArray(data)) return '';
            if (data.length === 0) return '';
            
            const headers = Object.keys(data[0]);
            const csvContent = [
                headers.join(','),
                ...data.map(row => headers.map(header => JSON.stringify(row[header] || '')).join(','))
            ].join('\n');
            
            return csvContent;
        }
        
        // Initialize data preview
        document.addEventListener('DOMContentLoaded', function() {
            const preview = document.getElementById('data-preview');
            if (Array.isArray(rawData) && rawData.length > 0) {
                const sample = rawData.slice(0, 3);
                preview.innerHTML = '<pre>' + JSON.stringify(sample, null, 2) + '</pre>';
            } else {
                preview.innerHTML = '<pre>' + JSON.stringify(rawData, null, 2) + '</pre>';
            }
        });
    </script>
</body>
</html>`

	tmpl, err := template.New("dashboard").Parse(dashboardTemplate)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	templateData := struct {
		Title       string
		CDNLibs     bool
		RecordCount int
		FieldCount  int
		LastUpdated string
		JSONData    string
	}{
		Title:       hg.Config.Title,
		CDNLibs:     hg.Config.CDNLibs,
		RecordCount: hg.getRecordCount(data),
		FieldCount:  hg.getFieldCount(data),
		LastUpdated: time.Now().Format("15:04:05"),
		JSONData:    string(jsonData),
	}

	htmlFile := filepath.Join(hg.Config.OutputDir, "index.html")
	file, err := os.Create(htmlFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, templateData)
}

func (hg *HTMLGenerator) generateTableView(data interface{}) error {
	tableTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Table View</title>
    {{if .CDNLibs}}
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.datatables.net/1.11.5/css/dataTables.bootstrap5.min.css" rel="stylesheet">
    <script src="https://code.jquery.com/jquery-3.6.0.min.js"></script>
    <script src="https://cdn.datatables.net/1.11.5/js/jquery.dataTables.min.js"></script>
    <script src="https://cdn.datatables.net/1.11.5/js/dataTables.bootstrap5.min.js"></script>
    {{end}}
</head>
<body>
    <div class="container-fluid">
        <nav class="navbar navbar-expand-lg navbar-dark bg-primary mb-4">
            <div class="container">
                <a class="navbar-brand" href="#">{{.Title}}</a>
                <ul class="navbar-nav">
                    <li class="nav-item"><a class="nav-link" href="index.html">Dashboard</a></li>
                    <li class="nav-item"><a class="nav-link active" href="table.html">Table View</a></li>
                    <li class="nav-item"><a class="nav-link" href="charts.html">Charts</a></li>
                    <li class="nav-item"><a class="nav-link" href="json.html">Raw JSON</a></li>
                </ul>
            </div>
        </nav>

        <div class="container-fluid">
            <div class="card">
                <div class="card-header">
                    <h5>Data Table</h5>
                </div>
                <div class="card-body">
                    <table id="data-table" class="table table-striped table-bordered" style="width:100%">
                        <thead id="table-header"></thead>
                        <tbody id="table-body"></tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>

    <script>
        const rawData = {{.JSONData}};
        
        document.addEventListener('DOMContentLoaded', function() {
            if (Array.isArray(rawData) && rawData.length > 0) {
                createDataTable(rawData);
            } else {
                document.getElementById('table-body').innerHTML = 
                    '<tr><td colspan="100%">No tabular data available</td></tr>';
            }
        });
        
        function createDataTable(data) {
            const headers = Object.keys(data[0]);
            
            // Create header
            const headerRow = document.getElementById('table-header');
            const headerRowHTML = '<tr>' + headers.map(h => '<th>' + h + '</th>').join('') + '</tr>';
            headerRow.innerHTML = headerRowHTML;
            
            // Create body
            const tbody = document.getElementById('table-body');
            const bodyHTML = data.map(row => 
                '<tr>' + headers.map(h => '<td>' + formatCellValue(row[h]) + '</td>').join('') + '</tr>'
            ).join('');
            tbody.innerHTML = bodyHTML;
            
            // Initialize DataTable
            $('#data-table').DataTable({
                pageLength: 25,
                responsive: true,
                order: [],
                columnDefs: [
                    { targets: '_all', className: 'text-nowrap' }
                ]
            });
        }
        
        function formatCellValue(value) {
            if (value === null || value === undefined) return '';
            if (typeof value === 'object') return JSON.stringify(value);
            return String(value);
        }
    </script>
</body>
</html>`

	tmpl, err := template.New("table").Parse(tableTemplate)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	templateData := struct {
		Title    string
		CDNLibs  bool
		JSONData string
	}{
		Title:    hg.Config.Title,
		CDNLibs:  hg.Config.CDNLibs,
		JSONData: string(jsonData),
	}

	htmlFile := filepath.Join(hg.Config.OutputDir, "table.html")
	file, err := os.Create(htmlFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, templateData)
}

func (hg *HTMLGenerator) generateChartView(data interface{}) error {
	chartTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Charts</title>
    {{if .CDNLibs}}
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    {{end}}
</head>
<body>
    <div class="container-fluid">
        <nav class="navbar navbar-expand-lg navbar-dark bg-primary mb-4">
            <div class="container">
                <a class="navbar-brand" href="#">{{.Title}}</a>
                <ul class="navbar-nav">
                    <li class="nav-item"><a class="nav-link" href="index.html">Dashboard</a></li>
                    <li class="nav-item"><a class="nav-link" href="table.html">Table View</a></li>
                    <li class="nav-item"><a class="nav-link active" href="charts.html">Charts</a></li>
                    <li class="nav-item"><a class="nav-link" href="json.html">Raw JSON</a></li>
                </ul>
            </div>
        </nav>

        <div class="container-fluid">
            <div class="row">
                <div class="col-md-6">
                    <div class="card">
                        <div class="card-header"><h6>Data Distribution</h6></div>
                        <div class="card-body">
                            <canvas id="distributionChart"></canvas>
                        </div>
                    </div>
                </div>
                <div class="col-md-6">
                    <div class="card">
                        <div class="card-header"><h6>Field Types</h6></div>
                        <div class="card-body">
                            <canvas id="typesChart"></canvas>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="row mt-4">
                <div class="col-12">
                    <div class="card">
                        <div class="card-header"><h6>Data Trends</h6></div>
                        <div class="card-body">
                            <canvas id="trendsChart"></canvas>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        const rawData = {{.JSONData}};
        
        document.addEventListener('DOMContentLoaded', function() {
            createCharts(rawData);
        });
        
        function createCharts(data) {
            if (Array.isArray(data) && data.length > 0) {
                createDistributionChart(data);
                createTypesChart(data);
                createTrendsChart(data);
            }
        }
        
        function createDistributionChart(data) {
            const ctx = document.getElementById('distributionChart').getContext('2d');
            new Chart(ctx, {
                type: 'doughnut',
                data: {
                    labels: ['Records', 'Fields', 'Empty Values'],
                    datasets: [{
                        data: [data.length, Object.keys(data[0] || {}).length, countEmptyValues(data)],
                        backgroundColor: ['#36A2EB', '#FFCE56', '#FF6384']
                    }]
                },
                options: { responsive: true }
            });
        }
        
        function createTypesChart(data) {
            if (data.length === 0) return;
            
            const typeCount = analyzeFieldTypes(data[0]);
            const ctx = document.getElementById('typesChart').getContext('2d');
            
            new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: Object.keys(typeCount),
                    datasets: [{
                        label: 'Field Count',
                        data: Object.values(typeCount),
                        backgroundColor: '#36A2EB'
                    }]
                },
                options: { 
                    responsive: true,
                    scales: { y: { beginAtZero: true } }
                }
            });
        }
        
        function createTrendsChart(data) {
            const ctx = document.getElementById('trendsChart').getContext('2d');
            new Chart(ctx, {
                type: 'line',
                data: {
                    labels: data.slice(0, 10).map((_, i) => 'Record ' + (i + 1)),
                    datasets: [{
                        label: 'Sample Data Trend',
                        data: data.slice(0, 10).map((_, i) => Math.random() * 100),
                        borderColor: '#36A2EB',
                        fill: false
                    }]
                },
                options: { responsive: true }
            });
        }
        
        function countEmptyValues(data) {
            return data.reduce((count, row) => {
                return count + Object.values(row).filter(v => v === null || v === undefined || v === '').length;
            }, 0);
        }
        
        function analyzeFieldTypes(row) {
            const types = {};
            Object.values(row).forEach(value => {
                const type = typeof value;
                types[type] = (types[type] || 0) + 1;
            });
            return types;
        }
    </script>
</body>
</html>`

	tmpl, err := template.New("charts").Parse(chartTemplate)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	templateData := struct {
		Title    string
		CDNLibs  bool
		JSONData string
	}{
		Title:    hg.Config.Title,
		CDNLibs:  hg.Config.CDNLibs,
		JSONData: string(jsonData),
	}

	htmlFile := filepath.Join(hg.Config.OutputDir, "charts.html")
	file, err := os.Create(htmlFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, templateData)
}

func (hg *HTMLGenerator) generateJSONView(data interface{}) error {
	jsonTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Raw JSON</title>
    {{if .CDNLibs}}
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    {{end}}
    <style>
        .json-container { max-height: 80vh; overflow-y: auto; }
        .json-view { background: #f8f9fa; padding: 15px; border-radius: 5px; }
        pre { margin: 0; white-space: pre-wrap; }
    </style>
</head>
<body>
    <div class="container-fluid">
        <nav class="navbar navbar-expand-lg navbar-dark bg-primary mb-4">
            <div class="container">
                <a class="navbar-brand" href="#">{{.Title}}</a>
                <ul class="navbar-nav">
                    <li class="nav-item"><a class="nav-link" href="index.html">Dashboard</a></li>
                    <li class="nav-item"><a class="nav-link" href="table.html">Table View</a></li>
                    <li class="nav-item"><a class="nav-link" href="charts.html">Charts</a></li>
                    <li class="nav-item"><a class="nav-link active" href="json.html">Raw JSON</a></li>
                </ul>
            </div>
        </nav>

        <div class="container-fluid">
            <div class="card">
                <div class="card-header d-flex justify-content-between">
                    <h5>Raw JSON Data</h5>
                    <button class="btn btn-sm btn-outline-primary" onclick="copyToClipboard()">Copy</button>
                </div>
                <div class="card-body json-container">
                    <div class="json-view">
                        <pre id="json-content">{{.JSONData}}</pre>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        function copyToClipboard() {
            const content = document.getElementById('json-content').textContent;
            navigator.clipboard.writeText(content).then(() => {
                alert('JSON copied to clipboard!');
            });
        }
    </script>
</body>
</html>`

	tmpl, err := template.New("json").Parse(jsonTemplate)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	templateData := struct {
		Title    string
		CDNLibs  bool
		JSONData string
	}{
		Title:    hg.Config.Title,
		CDNLibs:  hg.Config.CDNLibs,
		JSONData: string(jsonData),
	}

	htmlFile := filepath.Join(hg.Config.OutputDir, "json.html")
	file, err := os.Create(htmlFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, templateData)
}

func (hg *HTMLGenerator) generateStaticAssets() error {
	// Generate custom CSS file
	cssContent := `
/* Custom styles for API viewer */
.dashboard-card { transition: transform 0.2s; }
.dashboard-card:hover { transform: translateY(-2px); }
.metric-value { animation: countUp 0.5s ease-out; }

@keyframes countUp {
    from { opacity: 0; transform: scale(0.5); }
    to { opacity: 1; transform: scale(1); }
}

.table-responsive { border-radius: 8px; overflow: hidden; }
.card { box-shadow: 0 2px 4px rgba(0,0,0,0.1); border: none; }
.card-header { background: linear-gradient(45deg, #007bff, #0056b3); color: white; }

/* Dark mode support */
@media (prefers-color-scheme: dark) {
    body { background-color: #121212; color: #ffffff; }
    .card { background-color: #1e1e1e; }
    .json-view { background-color: #2d2d2d !important; }
}
`

	cssFile := filepath.Join(hg.Config.OutputDir, "styles.css")
	return os.WriteFile(cssFile, []byte(cssContent), 0644)
}

func (hg *HTMLGenerator) getRecordCount(data interface{}) int {
	if reflect.TypeOf(data).Kind() == reflect.Slice {
		return reflect.ValueOf(data).Len()
	}
	return 1
}

func (hg *HTMLGenerator) getFieldCount(data interface{}) int {
	if reflect.TypeOf(data).Kind() == reflect.Slice {
		slice := reflect.ValueOf(data)
		if slice.Len() > 0 {
			first := slice.Index(0)
			if first.Kind() == reflect.Map {
				return first.Len()
			}
		}
	}
	if reflect.TypeOf(data).Kind() == reflect.Map {
		return reflect.ValueOf(data).Len()
	}
	return 0
}