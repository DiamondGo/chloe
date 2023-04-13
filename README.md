Chloe is a telegram chat bot backended by openai.

You can talk to her 1 on 1 or in group chat. In group chat you need to name her at the beginning of your question or @ her bot username.

Please change config.yml and put your own keys there.

You need to get 4 things done to make this work:
* An OpenAI apikey: get one free at https://platform.openai.com/account .
* A telegram bot: send /newbot to BotFather on telegram.
* (optional) An api key on https://detectlanguage.com/ for free, if you want to produce text-to-speech synthesis voice.

Run command:  
go run main.go  
or:  
go build && ./chloe  

Run with docker:
* in docker dir
* modify your .env file for detectlanguage apikey
* docker-compose up
  
  
Have fun!

Ask question in text:  
![ask question in text](https://github.com/DiamondGo/blob/blob/chloe/ask_coding.jpg?raw=true)


Ask question in voice(if you want a reply in voice you need to setup pyservice localy):  
![ask question in text](https://github.com/DiamondGo/blob/blob/chloe/tts.jpg?raw=true)

Draw picture(Dall-e 2):  
![ask question in text](https://github.com/DiamondGo/blob/blob/chloe/draw_pic.jpg?raw=true)
