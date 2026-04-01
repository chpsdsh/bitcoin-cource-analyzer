package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Category      string `json:"category"`
	Summarization string `json:"summarization"`
	Features      struct {
		SignalDirection   string  `json:"signal_direction"`
		SignalStrength    float64 `json:"signal_strength"`
		Uncertainty       float64 `json:"uncertainty"`
		EventUrgencyHours int     `json:"event_urgency_hours"`
		NumbersDensity    float64 `json:"numbers_density"`
		EntityDensity     float64 `json:"entity_density"`
	} `json:"features"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))

	case http.MethodGet:
		resp := Response{
			Category:      "Macro",
			Summarization: "Федеральная резервная система сигнализирует о сохранении высоких процентных ставок на более длительный срок из-за устойчивой инфляции. Сильные данные по рынку труда США усилили ожидания ужесточения денежно-кредитной политики. Это оказывает давление на рискованные активы, включая Bitcoin, так как более высокие ставки снижают привлекательность инвестиций в криптовалюты.",
		}

		resp.Features.SignalDirection = "down"
		resp.Features.SignalStrength = 0.75
		resp.Features.Uncertainty = 0.2
		resp.Features.EventUrgencyHours = 24
		resp.Features.NumbersDensity = 0.6
		resp.Features.EntityDensity = 0.5

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
