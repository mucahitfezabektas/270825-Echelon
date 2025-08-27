package progress

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

var clientConnections = make(map[string]*websocket.Conn)
var mu sync.Mutex

func WebSocketHandler(c *websocket.Conn) {
	processID := c.Query("process_id")
	if processID == "" {
		log.Println("❌ WebSocket: process_id belirtilmedi.")
		c.WriteMessage(websocket.TextMessage, []byte("Hata: process_id belirtilmedi."))
		c.Close()
		return
	}

	mu.Lock()
	clientConnections[processID] = c
	mu.Unlock()

	log.Printf("✅ WebSocket: Yeni bağlantı kuruldu. Process ID: %s", processID)

	go func() {
		defer func() {
			mu.Lock()
			delete(clientConnections, processID)
			mu.Unlock()
			c.Close()
			// Bu log artık daha az "hata" gibi duracak, bağlantının kapandığını belirtecek.
			log.Printf("🗑️ WebSocket: Bağlantı temizlendi. Process ID: %s (Ping rutini sonlandı)", processID)
		}()

		for {
			time.Sleep(10 * time.Second)
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				// Ping hatası, bağlantının koptuğunu gösterir, bir hata değil, bir durumdur.
				// Log seviyesini düşürebilir veya sadece bir bilgi mesajı basabiliriz.
				log.Printf("ℹ️ WebSocket: Ping gönderilemedi (bağlantı kapalı/kopuk). Process ID: %s, Hata: %v", processID, err)
				return // Goroutine'den çık
			}
		}
	}()

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				// Bu bir hata değil, istemcinin bağlantıyı normal şekilde kapattığı anlamına gelir.
				log.Printf("ℹ️ WebSocket: İstemci bağlantısı normal şekilde kapandı. Process ID: %s", processID)
			} else {
				// Diğer beklenmedik hatalar için hala hata olarak logla.
				log.Printf("❌ WebSocket: Mesaj okuma hatası (beklenmedik), Process ID: %s, Hata: %v", processID, err)
			}
			break
		}
	}
}

func SendProgressUpdate(processID string, progressPercent int, message string) {
	mu.Lock()
	conn, ok := clientConnections[processID]
	mu.Unlock()

	if !ok {
		return
	}

	update := fmt.Sprintf(`{"progress": %d, "message": "%s"}`, progressPercent, message)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(update)); err != nil {
		// Mesaj gönderme hatası da bağlantının koptuğunu gösterebilir, önem düzeyini düşürebiliriz.
		log.Printf("ℹ️ WebSocket: İlerleme güncellenemedi (bağlantı kapalı/kopuk). Process ID: %s, Hata: %v", processID, err)
		mu.Lock()
		delete(clientConnections, processID)
		mu.Unlock()
		conn.Close()
	}
}
