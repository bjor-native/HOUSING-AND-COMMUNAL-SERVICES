package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const settingsFileName = "settings.json"

func main() {
	app()
}

func app() {
	data := readFile()

	settings := chooseHouse(data)
	fmt.Println("--------------------Информация o кваритире---------------------")
	fmt.Println("Квартира: " + settings.HouseName)
	fmt.Println("Тариф горячей воды: " + fmt.Sprint(settings.HeatWaterTariff))
	fmt.Println("Тариф холодной воды: " + fmt.Sprint(settings.ColdWaterTariff))
	fmt.Println("Тариф канализации: " + fmt.Sprint(settings.BothWaterTariff))
	if len(settings.History) == 0 {
		fmt.Println("Старых показаний нет.")
	}
	fmt.Println("----------------------История показаний------------------------")
	for _, history := range settings.History {
		timeHistory := time.Unix(history.Date, 0)
		fmt.Printf("Дата: Год - %v, месяц - %v, день - %v\n", timeHistory.Year(), timeHistory.Month(), timeHistory.Day())
		fmt.Printf("ГВС: %v\n", history.HotWater)
		fmt.Printf("XBC: %v\n", history.ColdWater)
		fmt.Println("---------------------------------------------------------------")
	}
	fmt.Println()
	fmt.Println("-----------")
	fmt.Println("|1. Воду  |")
	fmt.Println("|2. Свет  |")
	fmt.Println("|3. Выйти |")
	fmt.Println("-----------")

	reader := bufio.NewReader(os.Stdin)
	text := parseInput(reader, "Введите номер пункта: ")

	number := convertTextToInt(text)

	switch number {
	case 1:
		var oldHeatValue = 0
		var oldColdValue = 0
		if len(settings.History) == 0 {
			readerOldHeat := bufio.NewReader(os.Stdin)
			oldHeatValue = convertTextToInt(parseInput(readerOldHeat, "Введите старое показание горячей воды: "))

			readerCurrentlyHeat := bufio.NewReader(os.Stdin)
			oldColdValue = convertTextToInt(parseInput(readerCurrentlyHeat, "Введите старое показание холодной воды: "))
		} else {
			var lastValues = settings.History[len(settings.History)-1]
			oldHeatValue = lastValues.HotWater
			oldColdValue = lastValues.ColdWater
		}

		readerOldCold := bufio.NewReader(os.Stdin)
		textCurrentlyHeat := parseInput(readerOldCold, "Введите текущее показание горячей воды: ")

		readerCurrentlyCold := bufio.NewReader(os.Stdin)
		textCurrentlyCold := parseInput(readerCurrentlyCold, "Введите текущее показание холодной воды: ")

		currentlyHeatValue := convertTextToInt(textCurrentlyHeat)
		currentlyColdValue := convertTextToInt(textCurrentlyCold)
		resultWater := getWaterSums(oldHeatValue, currentlyHeatValue, oldColdValue, currentlyColdValue, settings)

		res := fmt.Sprintf("Сумма за воду в этом месяце равна: %.2f рублей.", resultWater)
		fmt.Println("------------------------------------------------------------------------")
		fmt.Println(res)
		fmt.Println("------------------------------------------------------------------------")
		app()
	case 2:
		readerOld := bufio.NewReader(os.Stdin)
		textOld := parseInput(readerOld, "Введите старое показание T общее: ")

		readerCurrently := bufio.NewReader(os.Stdin)
		textCurrently := parseInput(readerCurrently, "Введите текущее показание T общее: ")

		oldValue := convertTextToInt(textOld)
		currentlyValue := convertTextToInt(textCurrently)
		resLights := getLightSums(oldValue, currentlyValue)
		res := fmt.Sprintf("Сумма за свет в этом месяце равна: %.2f рублей.", resLights)
		fmt.Println("------------------------------------------------------------------------")
		fmt.Println(res)
		fmt.Println("------------------------------------------------------------------------")
		app()
	case 3:
		os.Exit(0)
	default:
		fmt.Println("------------------------------------------------------------------------")
		fmt.Println("Invalid")
		fmt.Println("------------------------------------------------------------------------")
		app()
	}
}

func chooseHouse(data HouseModel) HouseSettings {
	fmt.Println("Выберите квартиру: ")
	for i, house := range data.Houses {
		fmt.Println(fmt.Sprint(i+1) + ". " + house.HouseName)
	}

	i := len(data.Houses) + 1
	fmt.Println(fmt.Sprint(i) + ". Выход")

	reader := bufio.NewReader(os.Stdin)
	text := parseInput(reader, "Введите номер пункта: ")

	number := convertTextToInt(text)

	if i == number {
		os.Exit(0)
	}

	switch {
	case contains(number, data.Houses):
		return data.Houses[number-1]
	default:
		return data.Houses[0]
	}
}

func contains(v int, a []HouseSettings) bool {
	for i := range a {
		if i == v {
			return true
		}
	}
	return false
}

func indexOf(v string, a []HouseSettings) int {
	for i, h := range a {
		if h.Id == v {
			return i
		}
	}
	return 0
}

func getWaterSums(heatOld int, heatCurrently int, coldOld int, coldCurrently int, settingHouse HouseSettings) float32 {
	heatSpent := heatCurrently - heatOld
	coldSpent := coldCurrently - coldOld
	bothSpent := heatSpent + coldSpent

	result := (float32(heatSpent) * settingHouse.HeatWaterTariff) + (float32(coldSpent) * settingHouse.ColdWaterTariff) + (float32(bothSpent) * settingHouse.BothWaterTariff)

	currentlyValues := HistoryValues{
		Date:      time.Now().Local().UTC().Unix(),
		HotWater:  heatCurrently,
		ColdWater: coldCurrently,
	}

	newSettings := HouseSettings{
		Id:              settingHouse.Id,
		HouseName:       settingHouse.HouseName,
		HeatWaterTariff: settingHouse.HeatWaterTariff,
		ColdWaterTariff: settingHouse.ColdWaterTariff,
		BothWaterTariff: settingHouse.BothWaterTariff,
		History:         append(settingHouse.History, currentlyValues),
	}

	writeFile(newSettings)

	return result
}

func getLightSums(oldValue int, currentlyValue int) float32 {
	const lightTariff = 6.43

	var result = float32(currentlyValue-oldValue) * lightTariff

	return result
}

func convertTextToInt(str string) int {
	number, err := strconv.Atoi(strings.TrimSpace(str))
	if err != nil {
		fmt.Println("------------------------------------------------------------------------")
		fmt.Println("Вы ввели не число или неправильное значение.")
		fmt.Println("------------------------------------------------------------------------")
		app()
	}

	return number
}

func convertTextToFloat(str string) float64 {
	number, err := strconv.ParseFloat(strings.TrimSpace(str), 64)
	if err != nil {
		fmt.Println("------------------------------------------------------------------------")
		fmt.Println("Вы ввели не число или неправильное значение.")
		fmt.Println("------------------------------------------------------------------------")
		app()
	}

	return number
}

func parseInput(reader *bufio.Reader, text string) string {
	fmt.Print(text)
	text, err := reader.ReadString('\n')
	if err != nil {
		app()
	}
	return text
}

func readFile() HouseModel {
	// open input file
	file, err := os.ReadFile(settingsFileName)
	if err != nil {
		fmt.Println("Вероятно нет файла конфигурации, создать файл?")
		fmt.Println("1. Да")
		fmt.Println("2. Нет")
		reader := bufio.NewReader(os.Stdin)
		text := parseInput(reader, "Введите номер пункта: ")

		switch convertTextToInt(text) {
		case 1:
			fmt.Println("Создаю файл конфигурации")
			readerName := bufio.NewReader(os.Stdin)
			textName := strings.TrimSpace(parseInput(readerName, "Введите название квартиры: "))

			readerHotTariff := bufio.NewReader(os.Stdin)
			textHotTariff := convertTextToFloat(parseInput(readerHotTariff, "Введите тариф горячей воды: "))

			readerColdTariff := bufio.NewReader(os.Stdin)
			textColdTariff := convertTextToFloat(parseInput(readerColdTariff, "Введите тариф холодной воды: "))

			readerBothTariff := bufio.NewReader(os.Stdin)
			textBothTariff := convertTextToFloat(parseInput(readerBothTariff, "Введите тариф канализации: "))

			houseSetting := HouseSettings{
				Id:              uuid.NewString(),
				HouseName:       textName,
				HeatWaterTariff: float32(textHotTariff),
				ColdWaterTariff: float32(textColdTariff),
				BothWaterTariff: float32(textBothTariff),
				History:         []HistoryValues{},
			}

			data := HouseModel{
				Houses: []HouseSettings{
					houseSetting,
				},
			}

			createFile(data)
			return data
		case 2:
			os.Exit(0)
		default:
			app()
		}
	}

	data := HouseModel{}

	_ = json.Unmarshal([]byte(file), &data)

	return data
}

func writeFile(houseSetting HouseSettings) {
	file, err := os.ReadFile(settingsFileName)
	if err != nil {
		panic(err)
	}

	data := HouseModel{}
	_ = json.Unmarshal([]byte(file), &data)

	(&data).Houses[indexOf(houseSetting.Id, data.Houses)] = houseSetting
	as_json, _ := json.MarshalIndent(data, "", "\t")
	os.WriteFile(settingsFileName, as_json, 0644)
}

func createFile(data HouseModel) {
	as_json, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		panic(err)
	}
	os.WriteFile(settingsFileName, as_json, 0644)
}

type HouseSettings struct {
	Id              string          `json:"id"`
	HouseName       string          `json:"houseName"`
	HeatWaterTariff float32         `json:"heatWaterTariff"`
	ColdWaterTariff float32         `json:"coldWaterTariff"`
	BothWaterTariff float32         `json:"bothWaterTariff"`
	History         []HistoryValues `json:"histories"`
}

type HistoryValues struct {
	Date      int64 `json:"date"`
	HotWater  int   `json:"hotWater"`
	ColdWater int   `json:"coldWater"`
}

type HouseModel struct {
	Houses []HouseSettings `json:"houses"`
}
