# ARPARI XML feed parser / transformer

Mezivrstva pro zpracování produktových, skladových a cenových feedů při migraci e-shopu ARPARI z PrestaShopu na Shoptet.

Projekt bude připravovat validní Shoptet XML výstupy z původního katalogu a dodavatelských feedů. Aktuální implementace je zatím minimální Go scaffold pro ověření lokálního běhu a Railway deploye.

Podrobné obecné zadání je v souboru [ZADANI.md](ZADANI.md).

## Aktuální stav

Implementováno:

- Go HTTP server bez externích závislostí,
- endpoint `GET /` s odpovědí `Hello world!`,
- endpoint `GET /healthz` s odpovědí `ok`,
- základní test handleru,
- `Dockerfile` pro Railway deployment,
- `railway.json` s Dockerfile builderem.

Feed parser framework, transformace dodavatelských XML a persistence posledních validních feedů budou doplněné v dalších krocích.

## Požadavky pro lokální vývoj

Stačí Go:

```bash
go version
```

Projekt aktuálně nepoužívá žádné další lokální nástroje typu `air`, `just` nebo Makefile.

## Lokální běh

```bash
go run ./cmd/server
```

Server poslouchá na portu z environment variable `PORT`. Pokud není nastavena, použije `8080`.

```bash
curl http://localhost:8080/
curl http://localhost:8080/healthz
```

Očekávané odpovědi:

```text
Hello world!
ok
```

## Testy

```bash
go test ./...
```

## Railway deployment

Repo je připravené na deploy přes Railway z Git repozitáře.

Railway použije:

- root-level `railway.json`,
- `Dockerfile`,
- environment variable `PORT`, kterou Railway nastavuje automaticky.

Aplikace v kontejneru poslouchá na:

```text
0.0.0.0:${PORT}
```

Po deployi ověř:

```text
https://<railway-domain>/
https://<railway-domain>/healthz
```

## Budoucí cílový tok

```text
vstupní feedy / export z původního e-shopu
-> parser / transformer
-> validní Shoptet XML
-> automatické importy v Shoptetu
```

Hlavní provozní pravidlo:

```text
Nikdy nepřepsat poslední funkční XML rozbitým feedem.
```

## Plánované zdroje dat

První fáze:

- původní PrestaShop export produktů, kategorií a vazeb,
- STIMA katalog: `https://www.stima.cz/userfiles/xml/ITTC_SHT_products.xml`,
- STIMA sklad: `https://www.stima.cz/userfiles/xml/ITTC_SHT_stock.xml`,
- STIMA sklad + ceny: `https://www.stima.cz/userfiles/xml/ITTC_SHT_stock_price.xml`,
- Autronic dostupnost: `https://autronic.cz/feeds/availability-feed.xml`,
- případně Autronic katalog: `https://autronic.cz/feeds/product-feed.xml`,
- SEGO: `https://segocz.cz/src/Frontend/Files/Feeds/Catalog/zbozi_123456.xml`,
- HON: `https://www.webshop.officepro-brno.cz/import/HONClientFeed/HONClientFeed.xml`.

Druhá fáze:

- Sakypaky: `https://www.sakypaky.cz/export/b2b_partners_cs.xml`,
- Dřevočal: `https://www.matrace-drevocal.cz/feed/`.

Poznámka k Autronic dostupnostnímu feedu: `HEAD` request vrací 404, ale běžný `GET` vrací XML. Parser proto musí používat `GET`.

## Dokumentace Shoptet

- [Shoptet XML specifikace](https://developers.shoptet.com/shoptet-tools/shoptet-xml-specification/)
- [Automatické importy produktů](https://podpora.shoptet.cz/automaticke-importy-produktu/)
- [Šablony variant](https://podpora.shoptet.cz/sablony-variant/)
- [Příplatkové parametry](https://podpora.shoptet.cz/priplatkove-parametry/)
