# link-availability-checker

Веб-сервер, позволяющий пользователям проверять доступность интернет-ресурсов (доменов) и получать отчеты в виде PDF

<details>
<summary><h3>ТЗ</h3></summary>

Необходимо написать веб-сервер, в который пользователь может отправлять ссылки на интернет-ресурсы, как по одной ссылке, так и сразу несколько ссылок. Сервис должен в ответ отправлять пользователю статусы этих интернет-ресурсов (доступен/недоступен) и номер присвоенный данному набору ссылок.
```json
{"links": ["google.com", "malformedlink.gg"]}
```  
```json
{"links":{"google.com": "available", "malformedlink.gg": "not avaliable"}, "links_num": 1}
```  
Так же пользователь может отправить запрос со списком номеров ранее отправленных ссылок (`links_num`), а сервис должен будет вернуть `.pdf ` файл с отчетом о статусе интернет-ресурсов, входящих в этот список.  
```json
{"links_list":[1, 2]}
```
```
(pdf)
```
**Важное условие**: в сервисе должен быть предусмотрен сценарий его перезагрузки/остановки, при этом, насколько это возможно, задачи, которые находятся в это время в обработке (загрузке) необходимо не потерять. Во время такой остановки, в сервис могут поступать новые задачи от пользователей.  
**Не нужно использовать внешнюю инфраструктуру**: Docker, DB, nginx и пр. Можно и нужно пользоваться стандартными практиками и паттернами, которые желательно описать в readme.md файле. Если возникают моменты, которые необходимо дополнительно сообщить, опиши их в readme.md  

</details>

<details>
<summary><h3>Описание реализации</h3></summary>

Проект построен по паттерну Controller–Service–Repository  
    - Controller – обработка HTTP-запросов  
    - Service – бизнес-логика  
    - Repository – методы для работы с хранилищем данных    

Используется Fx для упрощения прокидывания зависимостей и управления жизненным циклом приложения, в качестве веб-сервера выступает Gin, для генерации PDF-файлов – [fpdf](https://codeberg.org/go-pdf/fpdf), хранение данных на диске при указании не использовать БД реализовано крайне упрощенно - наборы ссылок хранятся строками в текстовом файле

Из стандартных практик и паттернов были использованы worker pool для обработки ссылок и persistent queue для очереди задач  
*Worker pool* спавнит необходимое количество число горутин, которые берут задания из persistent queue для асинхронной обработки ссылок в наборе  
*Persistent queue* хранит очередь задач в памяти, выгружает на диск при остановке приложения и загружает обратно при старте

При тестировании после отправки 100+ ссылок для проверки уткнулся в большие задержки (30-40 секунд на обработку набора), пошел искать узкие места  
    - Перед отправкой HTTP-запроса сделал проверку DNS-имени, что позволило отсеять невалидные ссылки без ожидания таймаута HTTP-клиента  
    - Переключился на использование HEAD вместо GET, т.к. не нужно ждать загрузки тела ответа  
    - Настроил жесткие таймауты для HTTP-клиента  
    - Для проверки DNS-имен стал использовать 1.1.1.1 вместо стандартного резолвера  
По итогу удалось снизить время обработки до 6-10 секунд на 300 ссылок (лучшее/худшее, среднее - 8 сек.)

Вес исполняемого файла – ~22 МБ (с -ldflags "-s -w", без – ~32 МБ), потребление памяти - 15 МБ при старте, <80 МБ при нагрузке, потребление CPU до 30% при нагрузке

</details>

<details>
<summary><h3>Запуск проекта</h3></summary>

Склонировать репозиторий
```bash
git clone https://github.com/itskoshkin/30.11.2025 && cd 30.11.2025
```
Подсунуть конфиг (отредактировать содержимое при необходимости)
```bash
mv example_config.yaml config.yaml
```
Собрать и запустить приложение
```bash
go build -o link-availability-checker cmd/main.go && ./link-availability-checker
```

**Запросы для тестирования**  
- Отправить 1 ссылку
    ```bash
    curl -X POST http://localhost:8080/api/v1/links/check -H "Content-Type: application/json" -d '{"links": ["google.com"]}'
    ```
    Ожидаемый ответ:
    ```json
    {"links":{"google.com":"available"},"links_num":1}
    ```
- Отправить несколько ссылок
    ```bash
    curl -X POST http://localhost:8080/api/v1/links/check -H "Content-Type: application/json" -d '{"links": ["google.com", "malformedlink.gg"]}'
    ```
    Ожидаемый ответ:
    ```json
    {"links":{"google.com":"available", "malformedlink.gg":"not available"},"links_num":2}
    ```
- Отправить много ссылок
    ```bash
    curl -X POST http://localhost:8080/api/v1/links/check -H "Content-Type: application/json" -d '{"links":["google.com","nosuchsite.con","yahoo.com","fake-domain.xyz","github.com","nonexistent123.site","reddit.com","invalid-domain.abc","stackoverflow.com","doesnotexist000.com","microsoft.com","fakewebsite.qwe","apple.com","notarealsite.asd","wikipedia.org","randomfake.domain","mozilla.org","nullsite.test","linkedin.com","ghostdomain.zzz","twitter.com","nonexistsite.xyz","netflix.com","invalid1234.site","spotify.com","fakepage.abc","instagram.com","nothinghere000.com","dropbox.com","fakesite.qwe","slack.com","notreal.zzz","zoom.us","nonexistentxyz.com","ebay.com","fake-testsite.abc","paypal.com","nosuchplace.test","bing.com","ghost123.site","adobe.com","nonexistent000.zzz","hulu.com","invalidlink.qwe","airbnb.com","fakedomain.abc","tripadvisor.com","nullpage.test","quora.com","nothinghere.xyz","imdb.com","randomsite.zzz","cnn.com","ghost1234.com","bbc.com","fake-page.test","forbes.com","nonexist0000.site","nytimes.com","invalid-domain.abc","washingtonpost.com","fakesite123.qwe","medium.com","notrealpage.zzz","stackoverflow.blog","nosuchdomain000.com","duckduckgo.com","randomfake.abc","techcrunch.com","nullweb.test","venturebeat.com","ghostsite.zzz","wired.com","fake0000.site","arstechnica.com","nonexistentpage.abc","twitch.tv","notareal000.com","discord.com","invalidsite.test","slack.com","fakewebpage.zzz","trello.com","ghostdomain123.qwe","asana.com","nonexistentpage.abc","notion.so","nothinghere000.test","airtable.com","randomlink.zzz","flickr.com","fake123.site","pinterest.com","notarealsite.abc","imgur.com","ghostpage.qwe","tumblr.com","nonexistent000.test","soundcloud.com","invalidlink.zzz","behance.net","fakesite123.com","dribbble.com","nosuchdomain000.test","canva.com","nullpage.zzz","zoominfo.com","nothinghere0000.site","yelp.com","fake000.abc","trip.com","nonexistentpage.qwe","booking.com","ghostsite.zzz","expedia.com","invalidpage.test","hotels.com","fakedomain000.com","skyscanner.net","notarealpage.zzz","kayak.com","nonexist0000.test","orbitz.com","randomsite.qwe","agoda.com","fake123.com","travelocity.com","ghostdomain000.zzz","lonelyplanet.com","nullpage.test","trivago.com","nothinghere000.abc","udemy.com","invalidsite.zzz","coursera.org","fakepage123.test","edx.org","nonexistent0000.com","khanacademy.org","ghostlink.zzz","codecademy.com","notrealpage.test","freecodecamp.org","fakedomain000.qwe","pluralsight.com","nothinghere123.abc","linkedin.com","nonexistpage.zzz","angel.co","fake0000.site","producthunt.com","ghostpage.qwe","kickstarter.com","invalidlink.abc","patreon.com","nonexistent000.zzz","indiegogo.com","fakepage123.com","medium.com","notreal000.test","substack.com","randomlink.zzz","tumblr.com","ghostsite123.qwe","wordpress.com","fakesite000.abc","wix.com","nullpage.test","squarespace.com","nothinghere123.zzz","weebly.com","nonexistentpage.qwe","giphy.com","fake123.com","tenor.com","ghostlink000.zzz","imgur.com","invalidpage.test","photobucket.com","fakedomain.qwe","deviantart.com","notrealpage000.abc","behance.net","randomsite.zzz","dribbble.com","ghostpage000.test","flickr.com","fake1234.com","pinterest.com","nothinghere000.zzz","500px.com","nonexistentpage.qwe","canva.com","invalidlink.abc","adobe.com","fakedomain000.test","photoshop.com","notarealpage.zzz","illustrator.com","ghostsite000.qwe","affinity.com","randompage.abc","coreldraw.com","nothinghere0000.test","blender.org","fake123.zzz","autodesk.com","nonexistpage000.com","unity.com","ghostlink123.test","unrealengine.com","invalidsite000.zzz","cryengine.com","fakedomain123.abc","godotengine.org","notreal0000.com","cnet.com","randomlink000.test","techradar.com","ghostpage123.zzz","tomshardware.com","fake000000.site","pcmag.com","nothinghere123.abc","arstechnica.com","nonexistentpage000.test","engadget.com","invalidlink000.zzz","thenextweb.com","fakedomain0000.com","mashable.com","ghostsite0000.test","gizmodo.com","notareal123.zzz","lifehacker.com","randompage000.qwe","wired.com","fake0000000.site","theverge.com","nothinghere00000.abc","huffpost.com","nonexistentpage0000.test","buzzfeed.com","ghostlink0000.zzz","vice.com","invalidsite123.abc","vulture.com","fakedomain00000.com","rottentomatoes.com","notreal00000.test","imdb.com","randomlink0000.zzz","metacritic.com","ghostpage00000.qwe","boxofficemojo.com","fake000000000.site","espn.com","nothinghere000000.abc","bleacherreport.com","nonexistentpage00000.test","nba.com","invalidlink00000.zzz","nfl.com","fakedomain000000.com","mlb.com","notreal000000.test","nhl.com","ghostsite000000.zzz","soccer.com","fake00000000.qwe","fifa.com","nothinghere0000000.abc","uefa.com","nonexistentpage000000.test","espncricinfo.com","ghostlink000000.zzz","cricbuzz.com","invalidsite0000000.com","bbc.co.uk","fakedomain0000000.test","theguardian.com","notreal0000000.zzz","telegraph.co.uk","randomlink000000.qwe","independent.co.uk","fake0000000000.site","dailymail.co.uk","nothinghere00000000.abc","mirror.co.uk","nonexistentpage0000000.test","sky.com","ghostpage0000000.zzz","itv.com","fake123000.qwe","channel4.com","notrealpage0000.abc","hulu.com","randomsite0000.test","netflix.com","ghostlink123000.zzz","primevideo.com","fakepage000000.qwe","disneyplus.com","nothinghere123000.abc","hbomax.com","nonexistent0000000.test","peacocktv.com","ghostsite123000.zzz","paramountplus.com","fake000000123.com","apple.com","notrealpage0000000.test","google.com","randomlink123000.zzz","yahoo.com","ghostpage0000000.qwe","bing.com","fake00000000000.site","ask.com","nothinghere000000000.abc","aol.com","nonexistentpage00000000.test","duckduckgo.com","invalidlink00000000.zzz","wolframalpha.com","fakedomain00000000.com","wolframalph.com","notreal00000000.test","calculator.com","ghostsite00000000.zzz","geeksforgeeks.org","fake0000000000.qwe","stackoverflow.com","nothinghere0000000000.abc","reddit.com","nonexistentpage000000000.test","quora.com","ghostlink000000000.zzz","medium.com","fakepage000000000.qwe","dev.to","notareal000000000.abc","hashnode.com","randomsite000000000.test","producthunt.com","ghostsite000000000.zzz","kickstarter.com","fake000000000000.site","patreon.com","nothinghere00000000000.abc","indiegogo.com","nonexistentpage0000000000.test","github.com","invalidlink0000000000.zzz","gitlab.com","fakedomain0000000000.com","bitbucket.org","notreal0000000000.test","sourceforge.net","ghostpage00000000000.qwe","codepen.io","fake0000000000000.site","jsfiddle.net","nothinghere000000000000.abc","codesandbox.io","nonexistentpage00000000000.test","replit.com","ghostlink000000000000.zzz","heroku.com","fakepage000000000000.qwe","netlify.com","notreal000000000000.abc","vercel.com","randomsite000000000000.test","digitalocean.com","ghostsite000000000000.zzz","aws.amazon.com","fake00000000000000.site","azure.microsoft.com","nothinghere0000000000000.abc","cloud.google.com","nonexistentpage000000000000.test","ibm.com","ghostlink0000000000000.zzz","oracle.com","invalidsite0000000000000.com"]}'
    ```
    Ожидаемый ответ:  
    *<3 hours later>*
    ```json
    {"links":["<очень длинный фрагмент>"],"links_num":3}
    ```
- Запросить PDF-отчет по номерам наборов ссылок
    ```bash
    curl -X POST http://localhost:8080/api/v1/links/get_report -H "Content-Type: application/json" -d '{"links_list":[1]}' --output report.pdf
    ```
    ```bash
    curl -X POST http://localhost:8080/api/v1/links/get_report -H "Content-Type: application/json" -d '{"links_list":[1,2,3,4]}' --output report_1-4.pdf
    ```

</details>

<details>
<summary><h3>Точки роста</h3></summary>

- Сделать API асинхронным – ручка, который принимает набор ссылок и сразу возвращает `task_id`, а клиент затем забирает результат через опрашивание (polling) или WebSocket/SSE
- Кэшировать результаты проверок и переиспользовать их при повторных запросах одного домена в течение небольшого времени
- Улучшить способ проверки доступности, возможно использовать надёжный сторонний API

</details>
