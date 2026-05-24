# ARPARI feed parser roadmap

Tento dokument popisuje dlouhodobý plán vývoje feed parseru pro migraci ARPARI z PrestaShopu na Shoptet.

Cíl je držet projekt KISS: malá Go aplikace, minimum závislostí, bezpečné publikování posledních validních XML feedů a postupné napojování dodavatelů.

## Stav projektu

Aktuálně hotovo:

- Go HTTP server bez externích závislostí.
- Railway deployment přes `Dockerfile`.
- Veřejné endpointy `/` a `/healthz`.
- Základní test server handleru.

Aktuálně mimo rozsah:

- reálné stahování dodavatelských XML,
- transformace do Shoptet XML,
- perzistence posledních validních feedů,
- status běhů,
- cron nebo scheduled jobs,
- mapování produktů a kategorií.

## M0: Deploy scaffold

Status: hotovo.

Cíl:

- ověřit, že repo lze deploynout na Railway,
- ověřit, že aplikace běží přes veřejnou HTTPS doménu,
- udržet základ bez frameworků a databáze.

Akceptační kritéria:

- `GET /` vrací `Hello world!`,
- `GET /healthz` vrací `ok`,
- `go test ./...` prochází,
- Railway deploy funguje přes Dockerfile.

## M1: Lokální feed framework

Cíl:

- přidat minimální interní framework pro generování feedů,
- zatím bez reálných dodavatelů,
- ověřit bezpečný publish flow na dummy XML feedu.

Rozsah:

- přidat `cmd/rebuild`,
- přidat interní rozhraní pro feed generator,
- přidat storage vrstvu pro zápis do `data/feeds`,
- přidat atomický publish proces:
  - zápis do temp souboru,
  - validace,
  - rename na veřejný feed,
- přidat dummy feed `hello.xml`,
- servírovat feed přes `GET /feeds/hello.xml`.

Akceptační kritéria:

- `go run ./cmd/rebuild --supplier hello` vytvoří `data/feeds/hello.xml`,
- `go run ./cmd/server` servíruje `GET /feeds/hello.xml`,
- rozbitý temp feed nepřepíše poslední validní feed,
- existují testy pro storage/publish logiku.

## M2: Status a provozní metadata

Cíl:

- mít přehled o posledních bězích feedů,
- připravit základ pro debugging v Railway logs a přes HTTP.

Rozsah:

- přidat status storage do `data/status.json`,
- přidat `GET /status`,
- ukládat pro každý feed:
  - čas posledního běhu,
  - stav `success` nebo `failed`,
  - počet zpracovaných položek,
  - počet přeskočených položek,
  - poslední chybu,
- logovat běhy přes `log/slog`.

Akceptační kritéria:

- po rebuild běhu se aktualizuje `data/status.json`,
- `GET /status` vrací JSON,
- chyba rebuild běhu je vidět ve statusu i v logu,
- poslední validní XML zůstane zachované při chybě.

## M3: Railway Volume a production storage

Cíl:

- zajistit, že poslední validní feedy přežijí redeploy,
- připravit aplikaci na reálné Shoptet importy.

Rozsah:

- používat env `DATA_DIR`, fallback lokálně `./data`,
- na Railway mountnout Volume na `/data`,
- ukládat feedy do `${DATA_DIR}/feeds`,
- ukládat status do `${DATA_DIR}/status.json`,
- doplnit README o Railway Volume setup.

Akceptační kritéria:

- lokálně aplikace funguje bez env konfigurace,
- na Railway aplikace používá `/data`,
- po redeploy zůstane poslední validní feed dostupný,
- chybějící storage adresáře se vytvoří automaticky.

## M4: Shoptet XML writer a základní validace

Cíl:

- mít společnou vrstvu pro generování Shoptet XML,
- nezačínat každý dodavatel vlastním XML string builderem.

Rozsah:

- přidat Shoptet XML writer pro MVP subset:
  - produktový kód,
  - název,
  - cena,
  - sklad,
  - dostupnost,
  - EAN,
  - varianty v minimálním rozsahu,
- escapovat XML přes standardní Go nástroje,
- přidat základní validační pravidla:
  - XML je well-formed,
  - feed není prázdný,
  - počet položek nepřekročí plánovaný limit,
  - produkt nepřekročí limit variant.

Akceptační kritéria:

- testy ověřují XML escaping,
- testy ověřují jednoduchý produkt a varianty,
- nevalidní XML se nepublikuje,
- validační chyby jsou čitelné ve statusu.

## M5: První reálný feed: STIMA stock

Cíl:

- napojit první nízkorizikový reálný dodavatelský feed,
- ověřit celý tok download -> parse -> transform -> publish.

Proč STIMA stock:

- je menší než katalog,
- řeší skladovou aktualizaci,
- má nižší riziko zásahu do katalogových dat.

Rozsah:

- stahovat `https://www.stima.cz/userfiles/xml/ITTC_SHT_stock.xml`,
- parsovat streamingově přes Go XML decoder,
- generovat Shoptet stock feed,
- vystavit `GET /feeds/stima-stock.xml`,
- přidat timeouty a rozumné HTTP chyby.

Akceptační kritéria:

- `go run ./cmd/rebuild --supplier stima-stock` stáhne a publikuje feed,
- výstup je dostupný přes `/feeds/stima-stock.xml`,
- při nedostupném zdroji zůstane poslední validní feed,
- status ukazuje počty položek a výsledek běhu.

## M6: STIMA stock + price

Cíl:

- přidat pravidelnou aktualizaci skladů a cen.

Rozsah:

- stahovat `https://www.stima.cz/userfiles/xml/ITTC_SHT_stock_price.xml`,
- transformovat pouze pole, která mají být bezpečně aktualizována,
- vystavit `GET /feeds/stima-stock-price.xml`,
- doplnit test fixtures pro typické položky.

Akceptační kritéria:

- feed se generuje samostatně od `stima-stock`,
- chyby jednoho STIMA feedu nerozbijí druhý,
- výstup neobsahuje katalogová pole, která nechceme přepisovat.

## M7: Mapování a ochrana původního katalogu

Cíl:

- zabránit duplicitám a nechtěnému přepisování katalogových dat.

Rozsah:

- definovat konfigurační formát pro mapování:
  - dodavatelský kód -> Shoptet/původní kód,
  - dodavatelská kategorie -> cílová Shoptet kategorie,
- načítat mapování z repa nebo z datového adresáře,
- validovat chybějící a duplicitní mapování,
- zavést pravidla, která pole smí dodavatel aktualizovat.

Akceptační kritéria:

- duplicitní kód je chyba konfigurace,
- chybějící povinné mapování je vidět ve statusu,
- existující produkt se nepřepíše citlivými katalogovými poli bez explicitního pravidla.

## M8: STIMA katalog

Cíl:

- napojit katalogový STIMA feed s produkty a variantami,
- ošetřit Shoptet limity.

Rozsah:

- stahovat `https://www.stima.cz/userfiles/xml/ITTC_SHT_products.xml`,
- transformovat katalog do Shoptet XML podle pravidel,
- řešit varianty a EAN,
- detekovat produkty nad limitem variant,
- vystavit `GET /feeds/stima-products.xml`.

Akceptační kritéria:

- produkt nad limitem 512 variant není bez kontroly publikován jako rozbitý feed,
- status ukáže přeskočené nebo problematické produkty,
- výstup respektuje mapování a nepřepisuje chráněná katalogová pole.

## M9: Cron a ruční rebuild v produkci

Cíl:

- spouštět feedy automaticky i ručně.

Rozsah:

- přidat chráněný HTTP endpoint pro rebuild:
  - `POST /internal/rebuild/{supplier}`,
  - autorizace přes env token,
- připravit Railway Scheduled Job nebo cron volání,
- dokumentovat doporučené intervaly pro jednotlivé feedy.

Akceptační kritéria:

- rebuild endpoint bez tokenu odmítá request,
- rebuild endpoint s tokenem spustí vybraný feed,
- plánované běhy jsou vidět ve statusu,
- současné běhy stejného feedu se navzájem nepřepíšou.

## M10: Další dodavatelé první fáze

Cíl:

- postupně napojit zbývající dodavatele z první fáze.

Pořadí:

1. Autronic dostupnost.
2. SEGO.
3. HON.
4. Autronic katalog pouze pokud bude opravdu potřeba.

Specifika:

- Autronic dostupnost musí používat `GET`, ne `HEAD`,
- Autronic katalog má přes 20 000 položek a bude vyžadovat filtrování nebo dělení,
- SEGO je Zboží.cz styl feedu,
- HON má vlastní XML strukturu.

Akceptační kritéria:

- každý dodavatel má samostatný výstupní endpoint,
- chyba jednoho dodavatele neovlivní ostatní feedy,
- každý dodavatel má test fixtures pro typické i chybové položky.

## M11: Produkční hardening

Cíl:

- snížit riziko tichých chyb v automatickém provozu.

Rozsah:

- přidat časové limity pro download a transformace,
- omezit velikost stahovaných feedů,
- přidat lepší validační report,
- přidat základní metriky do statusu,
- doplnit dokumentaci provozních zásahů.

Akceptační kritéria:

- zdrojový feed, který visí nebo je extrémně velký, nezablokuje aplikaci,
- status jasně ukáže poslední úspěšný a poslední neúspěšný běh,
- provozní dokumentace popíše, co dělat při chybě feedu.

## M12: Druhá fáze dodavatelů

Cíl:

- napojit dodavatele odložené mimo MVP.

Rozsah:

- Sakypaky,
- Dřevočal,
- ověřit varianty matrací a konfigurace,
- vyhodnotit, jestli je potřeba rozšířit model příplatkových parametrů.

Akceptační kritéria:

- nové feedy dodržují stejný publish/status/validation flow,
- zvláštní pravidla dodavatelů jsou zdokumentovaná,
- výstupy jsou připravené pro konzervativní Shoptet import.

## Otevřené otázky

- Jaký bude finální formát exportu původního PrestaShop katalogu?
- Kde bude dlouhodobě spravované mapování produktových kódů?
- Kde bude dlouhodobě spravované mapování kategorií?
- Budeme používat jen strukturální validaci XML, nebo i oficiální Shoptet Relax NG validaci?
- Které konkrétní katalogové atributy smí každý dodavatel aktualizovat?
- Bude potřeba administrační UI, nebo dlouhodobě stačí config soubory a status endpoint?

## Doporučený nejbližší krok

Začít milníkem M1:

```bash
go run ./cmd/rebuild --supplier hello
go run ./cmd/server
curl http://localhost:8080/feeds/hello.xml
```

Tím vznikne první skutečný feed pipeline bez rizika, že rovnou řešíme složitosti reálných dodavatelů.
