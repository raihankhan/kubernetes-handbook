package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type SolrConfig struct {
	XMLName           xml.Name            `xml:"solr"`
	SolrCloud         SolrCloud           `xml:"solrcloud"`
	ShardHandler      ShardHandlerFactory `xml:"shardHandlerFactory"`
	MaxBooleanClauses int                 `xml:"int"`
	AllowPaths        string              `xml:"str"`
	Metrics           Metrics             `xml:"metrics"`
}

type SolrCloud struct {
	Host                     string `xml:"str"`
	HostPort                 int    `xml:"int"`
	HostContext              string `xml:"str"`
	GenericCoreNodeNames     bool   `xml:"bool"`
	DistribUpdateSoTimeout   int    `xml:"int"`
	DistribUpdateConnTimeout int    `xml:"int"`
}

type ShardHandlerFactory struct {
	Name          string `xml:"name,attr"`
	Class         string `xml:"class,attr"`
	SocketTimeout int    `xml:"int"`
	ConnTimeout   int    `xml:"int"`
}

type Metrics struct {
	Enabled bool `xml:"enabled,attr"`
}

type dynamicValue struct {
	Value string `xml:",chardata"`
}

func (dv *dynamicValue) Int() (int, error) {
	if dv.Value == "" {
		return 0, nil
	}
	return strconv.Atoi(dv.Value)
}

func main() {
	xmlString := `<?xml version="1.0" encoding="UTF-8" ?>
<solr>
  <solrcloud>
    <str name="host">${host:}</str>
    <int name="hostPort">${solr.port.advertise:80}</int>
    <str name="hostContext">${hostContext:solr}</str>
    <bool name="genericCoreNodeNames">${genericCoreNodeNames:true}</bool>
    <int name="distribUpdateSoTimeout">${distribUpdateSoTimeout:600000}</int>
    <int name="distribUpdateConnTimeout">${distribUpdateConnTimeout:60000}</int>
  </solrcloud>
  <shardHandlerFactory name="shardHandlerFactory" class="HttpShardHandlerFactory">
    <int name="socketTimeout">${socketTimeout:600000}</int>
    <int name="connTimeout">${connTimeout:60000}</int>
  </shardHandlerFactory>
  <int name="maxBooleanClauses">${solr.max.booleanClauses:1024}</int>
  <str name="allowPaths">${solr.allowPaths:}</str>
  <metrics enabled="${metricsEnabled:true}"/>
</solr>`

	var config SolrConfig
	decoder := xml.NewDecoder(strings.NewReader(xmlString))

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Error reading XML token: %v\n", err)
			return
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "solrcloud":
				err := decoder.DecodeElement(&config.SolrCloud, &t)
				if err != nil {
					fmt.Printf("Error decoding solrcloud: %v\n", err)
					return
				}
			case "shardHandlerFactory":
				err := decoder.DecodeElement(&config.ShardHandler, &t)
				if err != nil {
					fmt.Printf("Error decoding shardHandlerFactory: %v\n", err)
					return
				}
			case "int":
				var attr dynamicValue
				err := decoder.DecodeElement(&attr, &t)
				if err != nil {
					fmt.Printf("Error decoding dynamic value: %v\n", err)
					return
				}
				value, err := attr.Int()
				if err != nil {
					fmt.Printf("Error converting dynamic value to int: %v\n", err)
					return
				}
				config.MaxBooleanClauses = value
			case "str":
				var attr dynamicValue
				err := decoder.DecodeElement(&attr, &t)
				if err != nil {
					fmt.Printf("Error decoding dynamic value: %v\n", err)
					return
				}
				config.AllowPaths = attr.Value
			case "metrics":
				err := decoder.DecodeElement(&config.Metrics, &t)
				if err != nil {
					fmt.Printf("Error decoding metrics: %v\n", err)
					return
				}
			}
		}
	}

	fmt.Printf("%+v\n", config)
}
