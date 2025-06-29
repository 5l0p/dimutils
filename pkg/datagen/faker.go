package datagen

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// Faker data arrays and generators
var (
	firstNames = []string{
		"James", "Mary", "John", "Patricia", "Robert", "Jennifer", "Michael", "Linda",
		"William", "Elizabeth", "David", "Barbara", "Richard", "Susan", "Joseph", "Jessica",
		"Thomas", "Sarah", "Christopher", "Karen", "Charles", "Nancy", "Daniel", "Lisa",
		"Matthew", "Betty", "Anthony", "Helen", "Mark", "Sandra", "Donald", "Donna",
		"Steven", "Carol", "Paul", "Ruth", "Andrew", "Sharon", "Kenneth", "Michelle",
		"Emma", "Oliver", "Ava", "Elijah", "Charlotte", "William", "Sophia", "James",
		"Amelia", "Benjamin", "Isabella", "Lucas", "Mia", "Henry", "Evelyn", "Alexander",
	}

	lastNames = []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
		"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas",
		"Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson", "White",
		"Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker", "Young",
		"Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores",
		"Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell",
	}

	companies = []string{
		"TechCorp", "DataSystems", "CloudWorks", "DigitalSolutions", "InnovateLabs",
		"FutureTech", "SmartSystems", "NextGen", "CyberWorks", "TechFlow",
		"DataDriven", "CloudFirst", "DevOps Inc", "AgileWorks", "ScaleUp",
		"Microservices Ltd", "API Gateway", "Container Co", "Kubernetes Inc", "Docker Hub",
		"AWS Solutions", "Google Cloud", "Microsoft Azure", "Oracle Systems", "IBM Watson",
		"Apple Inc", "Meta Platforms", "Netflix", "Spotify", "Uber Technologies",
	}

	domains = []string{
		"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "aol.com",
		"icloud.com", "protonmail.com", "company.com", "business.org", "enterprise.net",
		"tech.io", "startup.co", "innovation.ai", "digital.dev", "cloud.tech",
	}

	streets = []string{
		"Main St", "Oak Ave", "First St", "Second St", "Park Ave", "Elm St",
		"Washington St", "Maple Ave", "Cedar St", "Pine St", "Lake Ave", "Hill St",
		"Church St", "School St", "High St", "Water St", "Mill St", "Spring St",
		"River Rd", "Sunset Blvd", "Broadway", "Madison Ave", "Franklin St", "Jefferson St",
	}

	cities = []string{
		"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia",
		"San Antonio", "San Diego", "Dallas", "San Jose", "Austin", "Jacksonville",
		"Fort Worth", "Columbus", "Charlotte", "San Francisco", "Indianapolis", "Seattle",
		"Denver", "Washington", "Boston", "El Paso", "Nashville", "Detroit", "Oklahoma City",
	}

	states = []string{
		"AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", "FL", "GA",
		"HI", "ID", "IL", "IN", "IA", "KS", "KY", "LA", "ME", "MD",
		"MA", "MI", "MN", "MS", "MO", "MT", "NE", "NV", "NH", "NJ",
		"NM", "NY", "NC", "ND", "OH", "OK", "OR", "PA", "RI", "SC",
		"SD", "TN", "TX", "UT", "VT", "VA", "WA", "WV", "WI", "WY",
	}

	loremWords = []string{
		"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit",
		"sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore",
		"magna", "aliqua", "enim", "ad", "minim", "veniam", "quis", "nostrud",
		"exercitation", "ullamco", "laboris", "nisi", "aliquip", "ex", "ea", "commodo",
		"consequat", "duis", "aute", "irure", "in", "reprehenderit", "voluptate", "velit",
		"esse", "cillum", "fugiat", "nulla", "pariatur", "excepteur", "sint", "occaecat",
		"cupidatat", "non", "proident", "sunt", "culpa", "qui", "officia", "deserunt",
		"mollit", "anim", "id", "est", "laborum", "the", "quick", "brown", "fox",
		"jumps", "over", "lazy", "dog", "pack", "my", "box", "with", "five", "dozen",
	}

	tlds = []string{
		".com", ".org", ".net", ".edu", ".gov", ".mil", ".int", ".io", ".co",
		".tech", ".dev", ".app", ".cloud", ".ai", ".ml", ".data", ".api",
	}
)

// generateString creates a random string based on configuration
func (g *Generator) generateString(config FieldConfig) string {
	if len(config.Values) > 0 {
		return config.Values[g.Rand.Intn(len(config.Values))]
	}
	
	if config.Pattern != "" {
		return g.generatePattern(config.Pattern)
	}
	
	minLen := 5
	maxLen := 20
	
	if config.Min != nil {
		if min, ok := config.Min.(float64); ok {
			minLen = int(min)
		}
	}
	if config.Max != nil {
		if max, ok := config.Max.(float64); ok {
			maxLen = int(max)
		}
	}
	
	length := minLen + g.Rand.Intn(maxLen-minLen+1)
	
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[g.Rand.Intn(len(charset))]
	}
	return string(b)
}

// generateInt creates a random integer
func (g *Generator) generateInt(config FieldConfig) int {
	min := 0
	max := 100
	
	if config.Min != nil {
		if minVal, ok := config.Min.(float64); ok {
			min = int(minVal)
		}
	}
	if config.Max != nil {
		if maxVal, ok := config.Max.(float64); ok {
			max = int(maxVal)
		}
	}
	
	if config.Distribution == "normal" {
		// Normal distribution around midpoint
		mean := float64(min+max) / 2
		stddev := float64(max-min) / 6
		value := g.Rand.NormFloat64()*stddev + mean
		
		if value < float64(min) {
			return min
		}
		if value > float64(max) {
			return max
		}
		return int(value)
	}
	
	return min + g.Rand.Intn(max-min+1)
}

// generateFloat creates a random float
func (g *Generator) generateFloat(config FieldConfig) float64 {
	min := 0.0
	max := 100.0
	
	if config.Min != nil {
		if minVal, ok := config.Min.(float64); ok {
			min = minVal
		}
	}
	if config.Max != nil {
		if maxVal, ok := config.Max.(float64); ok {
			max = maxVal
		}
	}
	
	return min + g.Rand.Float64()*(max-min)
}

// generateUUID creates a UUID-like string
func (g *Generator) generateUUID() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		g.Rand.Uint32(),
		g.Rand.Uint32()&0xffff,
		g.Rand.Uint32()&0xffff,
		g.Rand.Uint32()&0xffff,
		g.Rand.Uint64()&0xffffffffffff,
	)
}

// generateName creates a realistic full name
func (g *Generator) generateName() string {
	firstName := firstNames[g.Rand.Intn(len(firstNames))]
	lastName := lastNames[g.Rand.Intn(len(lastNames))]
	return fmt.Sprintf("%s %s", firstName, lastName)
}

// generateEmail creates a realistic email address
func (g *Generator) generateEmail() string {
	firstName := strings.ToLower(firstNames[g.Rand.Intn(len(firstNames))])
	lastName := strings.ToLower(lastNames[g.Rand.Intn(len(lastNames))])
	domain := domains[g.Rand.Intn(len(domains))]
	
	separator := ""
	if g.Rand.Intn(2) == 1 {
		separator = "."
	}
	
	return fmt.Sprintf("%s%s%s@%s", firstName, separator, lastName, domain)
}

// generatePhone creates a phone number
func (g *Generator) generatePhone() string {
	return fmt.Sprintf("(%03d) %03d-%04d",
		200+g.Rand.Intn(800),
		200+g.Rand.Intn(800),
		g.Rand.Intn(10000),
	)
}

// generateAddress creates a street address
func (g *Generator) generateAddress() string {
	number := 1 + g.Rand.Intn(9999)
	street := streets[g.Rand.Intn(len(streets))]
	city := cities[g.Rand.Intn(len(cities))]
	state := states[g.Rand.Intn(len(states))]
	zip := 10000 + g.Rand.Intn(90000)
	
	return fmt.Sprintf("%d %s, %s, %s %d", number, street, city, state, zip)
}

// generateCompany creates a company name
func (g *Generator) generateCompany() string {
	return companies[g.Rand.Intn(len(companies))]
}

// generateURL creates a URL
func (g *Generator) generateURL() string {
	protocol := "https"
	if g.Rand.Intn(10) == 0 {
		protocol = "http"
	}
	
	subdomain := ""
	if g.Rand.Intn(3) == 0 {
		subdomains := []string{"www", "api", "app", "dev", "staging", "cdn"}
		subdomain = subdomains[g.Rand.Intn(len(subdomains))] + "."
	}
	
	domain := strings.ToLower(companies[g.Rand.Intn(len(companies))])
	tld := tlds[g.Rand.Intn(len(tlds))]
	
	path := ""
	if g.Rand.Intn(2) == 0 {
		paths := []string{"/api/v1", "/docs", "/dashboard", "/admin", "/user", "/products"}
		path = paths[g.Rand.Intn(len(paths))]
	}
	
	return fmt.Sprintf("%s://%s%s%s%s", protocol, subdomain, domain, tld, path)
}

// generateTimestamp creates a timestamp
func (g *Generator) generateTimestamp() string {
	now := time.Now()
	// Random time within the last year
	past := now.AddDate(-1, 0, 0)
	diff := now.Unix() - past.Unix()
	randomTime := past.Add(time.Duration(g.Rand.Int63n(diff)) * time.Second)
	
	return randomTime.Format(time.RFC3339)
}

// generateDate creates a date
func (g *Generator) generateDate() string {
	year := 2020 + g.Rand.Intn(5)
	month := 1 + g.Rand.Intn(12)
	day := 1 + g.Rand.Intn(28) // Safe for all months
	
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

// generateTime creates a time
func (g *Generator) generateTime() string {
	hour := g.Rand.Intn(24)
	minute := g.Rand.Intn(60)
	second := g.Rand.Intn(60)
	
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

// generateLorem creates Lorem Ipsum text
func (g *Generator) generateLorem(config FieldConfig) string {
	wordCount := 10
	
	if config.Min != nil {
		if min, ok := config.Min.(float64); ok {
			wordCount = int(min)
		}
	}
	if config.Max != nil {
		if max, ok := config.Max.(float64); ok {
			maxWords := int(max)
			if maxWords > wordCount {
				wordCount = wordCount + g.Rand.Intn(maxWords-wordCount+1)
			}
		}
	}
	
	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		words[i] = loremWords[g.Rand.Intn(len(loremWords))]
	}
	
	text := strings.Join(words, " ")
	
	// Capitalize first letter
	if len(text) > 0 {
		text = strings.ToUpper(text[:1]) + text[1:]
	}
	
	return text + "."
}

// generateIP creates an IP address
func (g *Generator) generateIP() string {
	// Generate a realistic private IP range
	ranges := []string{
		"192.168.%d.%d",
		"10.%d.%d.%d",
		"172.16.%d.%d",
	}
	
	pattern := ranges[g.Rand.Intn(len(ranges))]
	
	switch pattern {
	case "192.168.%d.%d":
		return fmt.Sprintf(pattern, g.Rand.Intn(256), 1+g.Rand.Intn(254))
	case "172.16.%d.%d":
		return fmt.Sprintf(pattern, g.Rand.Intn(16), 1+g.Rand.Intn(254))
	default:
		return fmt.Sprintf(pattern, g.Rand.Intn(256), g.Rand.Intn(256), 1+g.Rand.Intn(254))
	}
}

// generateMAC creates a MAC address
func (g *Generator) generateMAC() string {
	mac := make(net.HardwareAddr, 6)
	for i := range mac {
		mac[i] = byte(g.Rand.Intn(256))
	}
	// Set the local bit to make it a locally administered address
	mac[0] |= 2
	return mac.String()
}

// generatePattern generates a string based on a pattern
func (g *Generator) generatePattern(pattern string) string {
	result := ""
	
	for i := 0; i < len(pattern); i++ {
		char := pattern[i]
		switch char {
		case '#':
			result += fmt.Sprintf("%d", g.Rand.Intn(10))
		case 'A':
			result += string(rune('A' + g.Rand.Intn(26)))
		case 'a':
			result += string(rune('a' + g.Rand.Intn(26)))
		case '?':
			chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
			result += string(chars[g.Rand.Intn(len(chars))])
		default:
			result += string(char)
		}
	}
	
	return result
}

// GenerateDataset creates a complete dataset with relationships
func (g *Generator) GenerateDataset(templates map[string]*DataTemplate) (map[string][]map[string]interface{}, error) {
	datasets := make(map[string][]map[string]interface{})
	
	// Generate data for each template
	for name, template := range templates {
		records := make([]map[string]interface{}, g.Config.Count)
		
		for i := 0; i < g.Config.Count; i++ {
			records[i] = g.generateRecord(template)
		}
		
		datasets[name] = records
	}
	
	// Apply relationships
	for name, template := range templates {
		if len(template.Relations) > 0 {
			g.applyRelations(datasets, name, template)
		}
	}
	
	return datasets, nil
}

func (g *Generator) applyRelations(datasets map[string][]map[string]interface{}, sourceName string, template *DataTemplate) {
	sourceData := datasets[sourceName]
	
	for _, relation := range template.Relations {
		targetData, exists := datasets[relation.Target]
		if !exists {
			continue
		}
		
		switch relation.Type {
		case "one-to-many":
			g.applyOneToMany(sourceData, targetData, relation)
		case "many-to-one":
			g.applyManyToOne(sourceData, targetData, relation)
		case "one-to-one":
			g.applyOneToOne(sourceData, targetData, relation)
		}
	}
}

func (g *Generator) applyOneToMany(sourceData, targetData []map[string]interface{}, relation Relation) {
	for i, sourceRecord := range sourceData {
		if i < len(targetData) {
			targetData[i][relation.Field] = sourceRecord["id"]
		}
	}
}

func (g *Generator) applyManyToOne(sourceData, targetData []map[string]interface{}, relation Relation) {
	for _, sourceRecord := range sourceData {
		if len(targetData) > 0 {
			targetRecord := targetData[g.Rand.Intn(len(targetData))]
			sourceRecord[relation.Field] = targetRecord["id"]
		}
	}
}

func (g *Generator) applyOneToOne(sourceData, targetData []map[string]interface{}, relation Relation) {
	minLen := len(sourceData)
	if len(targetData) < minLen {
		minLen = len(targetData)
	}
	
	for i := 0; i < minLen; i++ {
		sourceData[i][relation.Field] = targetData[i]["id"]
	}
}