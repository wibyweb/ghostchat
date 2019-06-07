# ghostchat

Ghostchat v1.0

Description: Simple anonymous HTTP chat server, requires no client side scripts. Works well with users on ancient machines.

Author: Wiby Search Engine

License: GPL v2


A few extra files are needed - Get full source at http://wiby.me/chat/ghostchat.zip


Enter chat: http://serverIP:PORT/chat/ - Hit refresh for it to begin working. Default port is 4444. 


--------------------------------------------------
Admin Commands (Place admin IPs in 'adminip' file)
--------------------------------------------------

Close chat: /close

Open chat: /open

Ban user and delete their posts: /ban userID

Enable or clear chat log: /log

Disable and delete chat log: /nolog

Clear chat feed: /clear

Change message of the day: /motd message

Remove message of the day: /motd

--------------------------------------------------

Note: 

  Anything inside the chat folder is served publicly.
  
	If cursor does not appear on form after pressing send, press Tab.
  
	Set a unique 93-byte key in the file called 'key'. Three bytes are used per day to create a
  
	usually unique ID based on the last 3 numbers of each client IP, ignoring octets.
  
	Full logs containing client IPs are located inside 'adminlog'.
  
	USE THIS CHAT SERVER AT YOUR OWN RISK.