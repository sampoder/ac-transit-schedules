package main
	
import (
		"fmt"
		"os"
		"net/http"
		"io"
		"encoding/json"
		"bytes"
		"strings"
		"github.com/olekukonko/tablewriter"
		"github.com/mbndr/figlet4go"
		"golang.org/x/exp/slices"
)

func formatJSON(data []byte) string {
		var out bytes.Buffer
		err := json.Indent(&out, data, "", " ")

		if err != nil {
				fmt.Println(err)
		}

		d := out.Bytes()
		return string(d)
}

type PredictionsResponse struct {
		BusTimeResponse BusTimeResponse `json:"bustime-response"`
}

type BusTimeResponse struct {
		Prd []BusPrediction `json:"prd"`
}

type BusPrediction struct {
		Tmstmp           string `json:"tmstmp"`
		Typ              string `json:"typ"`
		Stpnm            string `json:"stpnm"`
		Stpid            string `json:"stpid"`
		Vid              string `json:"vid"`
		Dstp             int    `json:"dstp"`
		Rt               string `json:"rt"`
		Rtdd             string `json:"rtdd"`
		Rtdir            string `json:"rtdir"`
		Des              string `json:"des"`
		Prdtm            string `json:"prdtm"`
		Tablockid        string `json:"tablockid"`
		Tatripid         string `json:"tatripid"`
		Origtatripno     string `json:"origtatripno"`
		Dly              bool   `json:"dly"`
		Dyn              int    `json:"dyn"`
		Prdctdn          string `json:"prdctdn"`
		Zone             string `json:"zone"`
		Rid              string `json:"rid"`
		Tripid           string `json:"tripid"`
		Tripdyn          int    `json:"tripdyn"`
		Schdtm           string `json:"schdtm"`
		Geoid            string `json:"geoid"`
		Seq              int    `json:"seq"`
		Psgld            string `json:"psgld"`
		Stst             int    `json:"stst"`
		Stsd             string `json:"stsd"`
		Flagstop         int    `json:"flagstop"`
}

type StopInformation struct {
		PlaceDescription string `json:"PlaceDescription"`
		Street           string `json:"Street"`
}

func padArrays(tableData [][]string) [][]string {
		if len(tableData) == 0 {
				return tableData
		}

		maxLength := len(tableData[0])
		for _, row := range tableData {
				if len(row) > maxLength {
						maxLength = len(row)
				}
		}
		
		for i, row := range tableData {
				for len(row) < maxLength {
						row = append(row, "")
						tableData[i] = row
				}
		}

		return tableData
}



func main() {
	
	if(!slices.Contains(os.Args[1:], "--no-title")){
		fmt.Println("\n")
		ascii := figlet4go.NewAsciiRender()
		renderStr, _ := ascii.Render("Bus Times")
		fmt.Print(renderStr)
	}
	
	for _, stop := range os.Args[1:] {
		
		if(stop == "--no-title"){
			return
		}
		
		client := &http.Client{}
	
		predictionsApiUrl := "https://api.actransit.org/transit/actrealtime/prediction?stpid="+ stop +"&token=TOKEN"
		
		predictionsRequest, _ := http.NewRequest("GET", predictionsApiUrl, nil)
		
		predictionsRequest.Header.Set("Content-Type", "application/json; charset=utf-8")
		
		predictionsResponse, _ := client.Do(predictionsRequest)
		
		predictionsResponseBody, _ := io.ReadAll(predictionsResponse.Body)
		
		defer predictionsResponse.Body.Close()
		
		stopApiUrl := "https://api.actransit.org/transit/stop/"+ stop +"/profile?token=TOKEN"
		
		stopRequest, _ := http.NewRequest("GET", stopApiUrl, nil)
		
		stopRequest.Header.Set("Content-Type", "application/json; charset=utf-8")
		
		stopResponse, _ := client.Do(stopRequest)
		
		stopResponseBody, _ := io.ReadAll(stopResponse.Body)
		
		defer stopResponse.Body.Close()
		
		var formattedPredictionsResponse PredictionsResponse
		
		json.Unmarshal([]byte(predictionsResponseBody), &formattedPredictionsResponse)
		
		var stopInformation StopInformation
		
		json.Unmarshal([]byte(stopResponseBody), &stopInformation)
		
		upcomingArrivals := make(map[string][]string)
		
		for _, prediction := range formattedPredictionsResponse.BusTimeResponse.Prd {
				key := prediction.Rt + " (" + prediction.Des + ")"
				value := strings.Split(prediction.Prdtm, " ")[1]
				if _, ok := upcomingArrivals[key]; ok {
						upcomingArrivals[key] = append(upcomingArrivals[key], value)
				} else {
						upcomingArrivals[key] = []string{value}
				}
		}
		
		tableData := [][]string{}
		
		for key, values := range upcomingArrivals {
				var arrival []string
				arrival = append(arrival, key)
				arrival = append(arrival, values...)
				tableData = append(tableData, arrival)
		}
		
		placeOrStreet := stopInformation.PlaceDescription
		if placeOrStreet == "" { placeOrStreet = stopInformation.Street }
		
		fmt.Println("\n \n", "🚏", placeOrStreet, "\n")
		
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Route"})
		table.SetCenterSeparator("|")
		table.AppendBulk(padArrays(tableData))
		table.Render()
	}
	
	fmt.Println("\n \n")
	
}