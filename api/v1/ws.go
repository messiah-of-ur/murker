package v1

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

// func ws(conn *websocket.Conn) {
// 	defer conn.Close()

// 	for {
// 		_, body, err := conn.ReadMessage()
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}

// 		log.Println(string(body))
// 	}
// }

// func handleWS(w http.ResponseWriter, r *http.Request) {
// 	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
// 	conn, err := upgrader.Upgrade(w, r, nil)

// 	if err != nil {
// 		log.Fatal(err)
// 		return
// 	}

// 	go ws(conn)
// }
