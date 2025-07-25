package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func SendPushNotification(pushToken, title, message string, data map[string]string) error {
	log.Printf("📩 Attempting to send push notification to: %s", pushToken)

	if len(pushToken) < 10 {
		log.Println("❌ Invalid Expo push token")
		return errors.New("invalid Expo push token")
	}
	if data == nil {
		data = map[string]string{}
	}
	pushMessage := map[string]interface{}{
		"to":    pushToken,
		"title": title,
		"body":  message,
		"sound": "default",
		"data":  data,
	}
	jsonData, err := json.Marshal([]map[string]interface{}{pushMessage})
	if err != nil {
		log.Printf("❌ Error serializing push message: %v", err)
		return err
	}

	expoAPI := "https://exp.host/--/api/v2/push/send"
	resp, err := http.Post(expoAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ HTTP request to Expo push service failed: %v", err)
		return err
	}
	defer resp.Body.Close()
	var responseMap map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseMap); err != nil {
		log.Printf("❌ Error decoding Expo response: %v", err)
		return err
	}

	log.Printf("📨 Expo Push Notification Response: %+v", responseMap)

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Expo push API error: %v", resp.Status)
		return errors.New("failed to send push notification")
	}

	log.Println("✅ Push notification sent successfully!")
	return nil
}

func ManualPushNotificationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		PushToken string `json:"pushToken"`
		Title     string `json:"title"`
		Message   string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.PushToken == "" {
		http.Error(w, "Push token is required", http.StatusBadRequest)
		return
	}

	err := SendPushNotification(req.PushToken, req.Title, req.Message, nil)
	if err != nil {
		log.Printf("❌ Failed to send push notification: %v", err)
		http.Error(w, "Failed to send notification", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Push notification sent successfully!"})
}
