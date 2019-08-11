//Ghostchat v1.0
//Description: Simple anonymous HTTP chat server, requires no client side scripts. Works well with users on ancient machines.
//Author: Wiby Search Engine
//License: GPL v2

//A few extra files are needed. Get full source at http://wiby.me/download/ghostchat.zip or https://github.com/wibyweb/ghostchat/

//Enter chat: http://serverIP:PORT/chat/ - Hit refresh for it to begin working. Default port is 4444. 

//------------------------------------------------------
//Admin Commands (Place admin IPs inside 'adminip' file)
//------------------------------------------------------
//Close chat: /close
//Open chat: /open
//Ban user and delete their posts: /ban userID
//Ban users and delete posts containing string: /banstr string
//Delete posts containing string: /delstr string
//Enable or clear chat log: /log
//Disable and delete chat log: /nolog
//Clear chat feed: /clear
//Change message of the day: /motd message
//Remove message of the day: /motd
//Filter swearwords: Add swearwords to 'swearfilter' file.
//------------------------------------------------------

//Note: Anything inside the chat folder is served publicly.
//	If cursor does not appear on form after pressing send, press Tab.
//	Set a unique 93-byte key in the file called 'key'. Three bytes are used per day to create a
//	usually unique ID based on the last 3 numbers of each client IP, ignoring octets.
//	Full logs containing client IPs are located inside 'adminlog'.
//	USE THIS CHAT SERVER AT YOUR OWN RISK.

package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"io/ioutil"
	"strings"
	"time"
	"strconv"
//	"encoding/hex"
	"math/big"
)

var feedheader string = "<html>\n<head>\n<meta http-equiv=\"refresh\" content=\"2\">\n</head>\n<body>\n"
var feedfooter string = "</body>\n</html>"

func main() {
	http.HandleFunc("/chat/post",handler)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":4444", nil))
//	log.Fatal(http.ListenAndServe("localhost:4444", nil))//For use behind a reverse proxy, make sure to send X-Real-IP header. Disable above line to use this one, also enable line 59 and disable line 60.
}

func handler(w http.ResponseWriter, r *http.Request) {
	//ip := r.Header.Get("X-Real-IP") //For use behind a reverse proxy. Make sure your reverse proxy is configured so header X-Real-IP cannot be spoofed. Also see comments on line 55
	ip := r.RemoteAddr[0:strings.LastIndex(r.RemoteAddr,":")]

	if checkban(ip) == false{
		switch r.Method {
			case "GET":
			http.Handle("/chat/", http.StripPrefix("/chat/", http.FileServer(http.Dir("chat/"))))

			case "POST":
			// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
			if err := r.ParseForm(); err != nil {
				fmt.Fprintf(w, "ParseForm() err: %v", err)
				return
			}

			formMessage := r.FormValue("message")

			//check if an admin command was made
			command := 0
			if formMessage != "" && formMessage[0] == '/'{ 
				command = commandhandler(formMessage, ip)
			}

			maxlines := 13 //max number of lines in feed. (13 is default for classic mac screen)

			//load contents of motd (message of the day)
			motdstring := ""
			motd, err := ioutil.ReadFile("motd")
			if err != nil {
				motdstring = ""
			}
			motdstring = string(motd)

			//load entire contents of feed
			feed, err := ioutil.ReadFile("chat/feed.html")
			if err != nil {
				panic(err)
			}
			feedstring := string(feed)

			if motdstring != "close\n" && motdstring != "close" && r.FormValue("message") != "" && command == 0{
				formMessageRune := []rune(formMessage)
				if len(formMessageRune) > 180{//trim posts greater than 180 runes
					formMessageRune = formMessageRune[0:179]
				}
				formMessage = string(formMessageRune)
				formMessage = strings.Replace(formMessage, "\n", "", -1)//user cant insert line feeds
				formMessage = swearfilter(formMessage)
				thetime := time.Now()
				//thetime.Format("2006-01-02 15:04:05")
				message := thetime.Format("15:04")
				message += " &lt;"
				message += createID(ip,thetime.Format("02"))
				message += "&gt; "
				message += formMessage
				message = strings.Replace(message, "<", "&lt;", -1)//user cant insert html
				message = strings.Replace(message, ">", "&gt;", -1)
				messagerune := []rune(message)

				//readjust lines in feed if motd present
				if motdstring != ""{
					motdrune := []rune(motdstring)
					if len(motdrune) > 95{
						maxlines -= 1
					}
					maxlines -= 1
				}

				//remove html header from feed
				feedstring = strings.Replace(feedstring, feedheader, "", -1)
				feedstring = strings.Replace(feedstring, feedfooter, "", -1)

				feedlinecount := 0
				messagelinecount := 0
				totalfeedlines := 0

				//posted message can take up to 2 lines
				if len(messagerune) > 90{
					messagelinecount += 2
				}else{
					messagelinecount ++
				}
				lines := strings.Split(feedstring, "<br>\n")
				feedtrimmed := ""
				rangecount := 1
				//find out how many lines there are currently in the feed 
				for _, line := range lines{
					//max 90 chars per line, and max message size is 180 or 2 lines
					linerune := []rune(line)
					if len(linerune) > 95{
						totalfeedlines ++
					}
					totalfeedlines ++
				}
				//trim old lines that wont fit once message is appended 
				for _, line := range lines{
					//max 90 chars per line, and max message size is 180 or 2 lines
					linerune := []rune(line)
					if len(linerune) > 95{
						feedlinecount ++
					}
					feedlinecount ++ //chat message will not exceed 180chars or 2 lines max

					if(feedlinecount > messagelinecount){
						if totalfeedlines - feedlinecount < maxlines {
							feedtrimmed += line
							if (len(lines) != rangecount){
								feedtrimmed += "<br>\n"
							}
						}
					}
					rangecount ++
				}
				//update the feed
				updatefeed := ""
				if (feedlinecount + messagelinecount > (maxlines+1)){
					feedtrimmed += message
					updatefeed += feedtrimmed
				}else{
					if motdstring != "" && feedstring != ""{//have to remove old motd line before appending it again
						motdend := strings.Index(feedstring,"<br>\n")
						feedstring = feedstring[motdend+5:len(feedstring)]
					}
					updatefeed += feedstring
					updatefeed += message
				}
				updatefeed += "<br>\n"
				updatedfeed := feedheader
				//if motd exists, append to feed. Make sure your motd isn't longer than 180 characters. 
				if motdstring != "" {
					updatedfeed += motdstring
					updatedfeed += "<br>\n"
				}
				updatedfeed += updatefeed
				updatedfeed += feedfooter

				//write feed
				f, err := os.OpenFile("chat/feed.html", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					panic(err)
				}
				defer f.Close()
				if _, err = f.WriteString(updatedfeed); err != nil {
					panic(err)
				}
				f.Sync()
				f.Close()

				//append log
				writelog(ip,message)

				//serve chat form again	
				http.ServeFile(w, r, "chat/form.html")
			}else if motdstring == "close\n" || motdstring == "close"{
				fmt.Fprintf(w, "The chat is currently closed.")
			}else{
				//serve chat form again	
				http.ServeFile(w, r, "chat/form.html")
			}

			default:
			fmt.Fprintf(w, "Only GET and POST is supported.")
		}
	}else{//Note that GET requests will still continue working until server is restarted or user restarts
		fmt.Fprintf(w, "403 Forbidden")
	}
}

func swearfilter(s string) string{
	swearwords := ""
	swearfilter, err := ioutil.ReadFile("swearfilter")
	if err != nil {
			swearwords = ""
	}else{
		swearwords = string(swearfilter)
	}
	swearwords = strings.Replace(swearwords, "\r", "", -1)//some editors like to insert carriage returns with line feeds
	if(len(swearwords) > 0){
		if swearwords[len(swearwords)-1] == byte('\n'){//some text editors like to do this at the final byte, can't have that
			swearwords = swearwords[0:len(swearwords)-1]
		}
		swears := strings.Split(swearwords, "\n")
		sLower := strings.ToLower(s)
		for _, swearword := range swears{
			if strings.Contains(sLower,swearword){
				return "I have a potty mouth."
			}
		}
	}
	return s
}

func checkban(ip string) bool{
	//fmt.Printf("\n%s",ip)
	//load entire contents of blockip
	blockip, err := ioutil.ReadFile("blockip")
	if err != nil {
		panic(err)
	}
	blockipstring := string(blockip)
	blockipstring = strings.Replace(blockipstring, "\r", "", -1)
	lines := strings.Split(blockipstring, "\n")
	for _, line := range lines{
		if ip == line{
			return true
		}
	}
	return false
}

func checkAdminIP(ip string) bool{
	adminip, err := ioutil.ReadFile("adminip")
	if err != nil {
		panic(err)
	}
	adminipstring := string(adminip)
	adminipstring = strings.Replace(adminipstring, "\r", "", -1)
	lines := strings.Split(adminipstring, "\n")
	for _, line := range lines{
		if ip == line{
			return true
		}
	}
	return false
}

func writelog(ip string, message string){
	logmessage := ip
	logmessage += " "
	logmessage += message
	logmessage += "\n"
	flog, err := os.OpenFile("adminlog", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer flog.Close()
	if _, err = flog.WriteString(logmessage); err != nil {
		panic(err)
	}
	flog.Sync()
	flog.Close()

	//appends to a public chat log (chatlog.html) but only if you create this file yourself
	fchatlog, err := os.OpenFile("chat/chatlog.html", os.O_APPEND|os.O_WRONLY, 0600)
	if err == nil {
		message += "<br>\n"
		defer fchatlog.Close()
		if _, err = fchatlog.WriteString(message); err != nil {
			panic(err)
		}
		fchatlog.Sync()
	}
	fchatlog.Close()
}

func commandhandler(formMessage string, ip string) int{
	noun := ""
	if len(formMessage) > 6 && formMessage[0:6] == "/motd "{
		noun = strings.TrimPrefix(formMessage, "/motd ")
		formMessage = "/motd"
	}
	if len(formMessage) > 5 && formMessage[0:5] == "/ban "{
		noun = strings.TrimPrefix(formMessage, "/ban ")
		formMessage = "/ban"
	}
	if len(formMessage) > 8 && formMessage[0:8] == "/banstr "{
		noun = strings.TrimPrefix(formMessage, "/banstr ")
		formMessage = "/banstr"
	}
	if len(formMessage) > 8 && formMessage[0:8] == "/delstr "{
		noun = strings.TrimPrefix(formMessage, "/delstr ")
		formMessage = "/delstr"
	}
	switch formMessage{ //process any verbs
	case "/close": //close the chat
		if checkAdminIP(ip) == true{
			motdstring := ""
			//read motd
			motd, err := ioutil.ReadFile("motd")
			if err == nil{
				motdstring = string(motd)
			}
			if motdstring != "close"{
				//backup motd
				motdbak, err := os.OpenFile("motdbak", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					panic(err)
				}
				defer motdbak.Close()
				if _, err = motdbak.WriteString(motdstring); err != nil {
					panic(err)
				}
				motdbak.Sync()
				motdbak.Close()

				//write "close" to motd
				motdupdate, err := os.OpenFile("motd", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					panic(err)
				}
				defer motdupdate.Close()
				if _, err = motdupdate.WriteString("close"); err != nil {
					panic(err)
				}
				motdupdate.Sync()
				motdupdate.Close()

				//update feed to stop the auto refresh
				feed, err := ioutil.ReadFile("chat/feed.html")
				if err != nil {
					panic(err)
				}
				feedstring := string(feed)
				feedstring = strings.Replace(feedstring, "http-equiv=\"refresh\" content=", "", -1)
				f, err := os.OpenFile("chat/feed.html", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					panic(err)
				}
				defer f.Close()
				if _, err = f.WriteString(feedstring); err != nil {
					panic(err)
				}
					f.Sync()
					f.Close()
				}
		}
		return 0
	case "/open": //open the chat
		if checkAdminIP(ip) == true{
			motdstring := ""
			//read motd
			motd, err := ioutil.ReadFile("motd")
			if err == nil{
				motdstring = string(motd)
			}
			if motdstring == "close"{
				//read motdbak
				motdbak, err := ioutil.ReadFile("motdbak")
				if err != nil {
					motdstring = ""
				}else{
					motdstring = string(motdbak)
				}

				//restore motd
				motdrestore, err := os.OpenFile("motd", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					panic(err)
				}
				defer motdrestore.Close()
				if _, err = motdrestore.WriteString(motdstring); err != nil {
					panic(err)
				}
				motdrestore.Sync()
				motdrestore.Close()
			}
		}
		return 0
	case "/log": //enable or clear chatlog
		if checkAdminIP(ip) == true{
			chatlog, err := os.OpenFile("chat/chatlog.html", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				panic(err)
			}
			chatlog.Close()
		}
		return 1
	case "/nolog": //delete and disable chatlog
		if checkAdminIP(ip) == true{
			err := os.Remove("chat/chatlog.html")
			if err != nil {
				fmt.Printf("\nChatlog already disabled.")
			}
		}
		return 1
	case "/clear": //clear the feed
		if checkAdminIP(ip) == true{
			feed, err := os.OpenFile("chat/feed.html", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				panic(err)
			}
			defer feed.Close()
			if _, err = feed.WriteString(feedheader + feedfooter); err != nil {
				panic(err)
			}
			feed.Sync()
			feed.Close()
		}
		return 1
	case "/motd": //update motd - not adding a message will remove motd
		if checkAdminIP(ip) == true{
			motdupdate, err := os.OpenFile("motd", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				panic(err)
			}
			defer motdupdate.Close()
			if _, err = motdupdate.WriteString(noun); err != nil {
				panic(err)
			}
			motdupdate.Sync()
			motdupdate.Close()
		}
		return 0
	case "/ban": //ban a user and remove their prior posts		
		if checkAdminIP(ip) == true && noun != ""{
			//read chat feed
			feedstring := ""
			feed, err := ioutil.ReadFile("chat/feed.html")
			if err == nil{
				feedstring = string(feed)
			}
			//read chatlog (if available)
			chatlogstring := ""
			chatlog, err := ioutil.ReadFile("chat/chatlog.html")
			if err == nil{
				chatlogstring = string(chatlog)
			}
			//read adminlog
			adminlogstring := ""
			adminlog, err := ioutil.ReadFile("adminlog")
			if err == nil{
				adminlogstring = string(adminlog)
			}

			//strip header from chat feed
			strings.TrimPrefix(feedstring, feedheader)
			//strip footer from chat feed
			strings.Replace(feedstring, feedfooter, "", -1)

			//loop over every line in feed, remove all lines of user
			cleansedfeed := feedheader
			feedlines := strings.Split(feedstring, "<br>\n")
			for _, line := range feedlines{
				linecpy := strings.Split(line, "&gt;")[0]//this does what "strings.trimright" should do but doesn't seem to do!
				if strings.Contains(linecpy,noun) == false{
					cleansedfeed += line
					cleansedfeed += "<br>\n"
				}
			}
			cleansedfeed += feedfooter

			//update feed
			f, err := os.OpenFile("chat/feed.html", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			if _, err = f.WriteString(cleansedfeed); err != nil {
				panic(err)
			}
			f.Sync()
			f.Close()

			//loop over every line in chatlog, remove all lines of user
			if chatlogstring != ""{
				cleansedchatlog := ""
				chatloglines := strings.Split(chatlogstring, "<br>\n")
				for _, line := range chatloglines{
					linecpy := strings.Split(line, "&gt;")[0]
					if strings.Contains(linecpy,noun) == false{
						cleansedchatlog += line
						cleansedchatlog += "<br>\n"
					}
				}

				//update chatlog
				cl, err := os.OpenFile("chat/chatlog.html", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					panic(err)
				}
				defer cl.Close()
				if _, err = cl.WriteString(cleansedchatlog); err != nil {
					panic(err)
				}
				cl.Sync()
				cl.Close()
			}

			//get ip of target user
			targetip := ""
			adminloglines := strings.Split(adminlogstring, "\n")
			for _, line := range adminloglines{
				line = strings.Split(line, "&gt;")[0]
				if strings.Contains(line,noun){
					targetip = strings.Split(line, " ")[0]
					targetip = strings.Replace(targetip, " ", "", -1)
				}
			}

			//append IP to blockip file
			if targetip != ""{
				blockip, err := os.OpenFile("blockip", os.O_APPEND|os.O_WRONLY, 0600)
				if err == nil {
					appendip := "\n"
					appendip += targetip
					defer blockip.Close()
					if _, err = blockip.WriteString(appendip); err != nil {
						panic(err)
					}
					blockip.Sync()
				}
				blockip.Close()
			}
		}
		return 1
	case "/banstr": //ban users that posted a specific string and delete the posts (botnet spam)		
		if checkAdminIP(ip) == true && noun != ""{
			//read chat feed
			feedstring := ""
			feed, err := ioutil.ReadFile("chat/feed.html")
			if err == nil{
				feedstring = string(feed)
			}
			//read chatlog (if available)
			chatlogstring := ""
			chatlog, err := ioutil.ReadFile("chat/chatlog.html")
			if err == nil{
				chatlogstring = string(chatlog)
			}
			//read adminlog
			adminlogstring := ""
			adminlog, err := ioutil.ReadFile("adminlog")
			if err == nil{
				adminlogstring = string(adminlog)
			}

			//strip header from chat feed
			strings.TrimPrefix(feedstring, feedheader)
			//strip footer from chat feed
			strings.Replace(feedstring, feedfooter, "", -1)

			//loop over every line in feed, remove all lines containing string
			cleansedfeed := feedheader
			feedlines := strings.Split(feedstring, "<br>\n")
			for _, line := range feedlines{
				if strings.Contains(strings.ToLower(line),strings.ToLower(noun)) == false{
					cleansedfeed += line
					cleansedfeed += "<br>\n"
				}
			}
			cleansedfeed += feedfooter

			//update feed
			f, err := os.OpenFile("chat/feed.html", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			if _, err = f.WriteString(cleansedfeed); err != nil {
				panic(err)
			}
			f.Sync()
			f.Close()

			//loop over every line in chatlog, remove all lines containing string
			if chatlogstring != ""{
				cleansedchatlog := ""
				chatloglines := strings.Split(chatlogstring, "<br>\n")
				for _, line := range chatloglines{
					if strings.Contains(strings.ToLower(line),strings.ToLower(noun)) == false{
						cleansedchatlog += line
						cleansedchatlog += "<br>\n"
					}
				}

				//update chatlog
				cl, err := os.OpenFile("chat/chatlog.html", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					panic(err)
				}
				defer cl.Close()
				if _, err = cl.WriteString(cleansedchatlog); err != nil {
					panic(err)
				}
				cl.Sync()
				cl.Close()
			}

			//get ips of users with target string
			targetips := ""
			targetip := ""
			adminloglines := strings.Split(adminlogstring, "\n")
			for _, line := range adminloglines{
				if strings.Contains(strings.ToLower(line),strings.ToLower(noun)){
					targetip = strings.Split(line, " ")[0]
					targetip = strings.Replace(targetip, " ", "", -1)
					if strings.Contains(targetips,targetip) == false{
						targetips += "\n"
						targetips += targetip
					}
				}
			}

			//append IPs to blockip file
			if targetips != ""{
				blockip, err := os.OpenFile("blockip", os.O_APPEND|os.O_WRONLY, 0600)
				if err == nil {
					defer blockip.Close()
					if _, err = blockip.WriteString(targetips); err != nil {
						panic(err)
					}
					blockip.Sync()
				}
				blockip.Close()
			}
		}
		return 1
	case "/delstr": //Delete posts that contain string		
		if checkAdminIP(ip) == true && noun != ""{
			//read chat feed
			feedstring := ""
			feed, err := ioutil.ReadFile("chat/feed.html")
			if err == nil{
				feedstring = string(feed)
			}
			//read chatlog (if available)
			chatlogstring := ""
			chatlog, err := ioutil.ReadFile("chat/chatlog.html")
			if err == nil{
				chatlogstring = string(chatlog)
			}

			//strip header from chat feed
			strings.TrimPrefix(feedstring, feedheader)
			//strip footer from chat feed
			strings.Replace(feedstring, feedfooter, "", -1)

			//loop over every line in feed, remove all lines containing string
			cleansedfeed := feedheader
			feedlines := strings.Split(feedstring, "<br>\n")
			for _, line := range feedlines{
				if strings.Contains(strings.ToLower(line),strings.ToLower(noun)) == false{
					cleansedfeed += line
					cleansedfeed += "<br>\n"
				}
			}
			cleansedfeed += feedfooter

			//update feed
			f, err := os.OpenFile("chat/feed.html", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			if _, err = f.WriteString(cleansedfeed); err != nil {
				panic(err)
			}
			f.Sync()
			f.Close()

			//loop over every line in chatlog, remove all lines containing string
			if chatlogstring != ""{
				cleansedchatlog := ""
				chatloglines := strings.Split(chatlogstring, "<br>\n")
				for _, line := range chatloglines{
					if strings.Contains(strings.ToLower(line),strings.ToLower(noun)) == false{
						cleansedchatlog += line
						cleansedchatlog += "<br>\n"
					}
				}

				//update chatlog
				cl, err := os.OpenFile("chat/chatlog.html", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					panic(err)
				}
				defer cl.Close()
				if _, err = cl.WriteString(cleansedchatlog); err != nil {
					panic(err)
				}
				cl.Sync()
				cl.Close()
			}
		}
		return 1
	default:
		return 0
	}
}

func createID(ip string, date string) string{
	//take last 3 numbers of client IP, xor it with a key. Just want to make a somewhat unique id that changes daily for users to call each other.
	ip = strings.Replace(ip, ".", "", -1)
	last3 := len(ip) - 3
	ip = ip[last3:len(ip)]
	//load key file - remember to set your own 93 byte key string instead of the default
	key, err := ioutil.ReadFile("key")
	if err != nil {
		panic(err)
	}
	dateint, _ := strconv.Atoi(date)
	keystring := string(key)
	keystring = keystring[dateint*3-3:dateint*3]//every day of the month gets a new 3 byte key, based on a key string of at least 93 bytes (3 bytes per day of the month up to 31 days)
	id := ""
	//xor ip with key
	for i := 0; i < len(ip); i++ {
			id += string(ip[i] ^ keystring[i])
	}
	//convert to a number system so ID will render properly
	//return hex.EncodeToString([]byte(id)) //going to use base36 instead of base16 for compactness
	idstring := string(EncodeBytesAsBytes([]byte(id)))
	idstring = idstring[1:len(idstring)]
	return idstring
}


//--------------------------------------------------------------------------------------------
//base36 encoder from https://github.com/martinlindhe/base36

var (
	base36 = []byte{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
		'u', 'v', 'w', 'x', 'y', 'z'}
)

// EncodeBytesAsBytes encodes a byte slice to base36.
func EncodeBytesAsBytes(b []byte) []byte {
	var bigRadix = big.NewInt(36)
	var bigZero = big.NewInt(0)
	x := new(big.Int)
	x.SetBytes(b)

	answer := make([]byte, 0, len(b)*136/100)
	for x.Cmp(bigZero) > 0 {
		mod := new(big.Int)
		x.DivMod(x, bigRadix, mod)
		answer = append(answer, base36[mod.Int64()])
	}

	// leading zero bytes
	for _, i := range b {
		if i != 0 {
			break
		}
		answer = append(answer, base36[0])
	}

	// reverse
	alen := len(answer)
	for i := 0; i < alen/2; i++ {
		answer[i], answer[alen-1-i] = answer[alen-1-i], answer[i]
	}

	return answer
}


