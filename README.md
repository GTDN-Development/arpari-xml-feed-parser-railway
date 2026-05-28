# ARPARI XML feed parser / transformer

Mezivrstva pro zpracování produktových, skladových a cenových feedů při migraci e-shopu ARPARI z PrestaShopu na Shoptet.

Projekt bude připravovat validní Shoptet XML výstupy z původního katalogu a dodavatelských feedů. Aktuální implementace je zatím minimální Go scaffold pro ověření lokálního běhu a Railway deploye.

Podrobné obecné zadání je v souboru [ZADANI.md](ZADANI.md).

## Aktuální stav

Implementováno:

- Go HTTP server bez externích závislostí,
- endpoint `GET /` s odpovědí `Hello world!`,
- endpoint `GET /healthz` s odpovědí `ok`,
- lokální dummy feed pipeline přes `cmd/rebuild`,
- endpoint `GET /feeds/hello.xml` pro vygenerovaný dummy XML feed,
- STIMA katalogový MVP feed `stima-products` z `ITTC_SHT_products.xml`,
- STIMA testovací katalog `stima-products-test` s prvními 20 produkty,
- STIMA skladový MVP feed `stima-stock` z `ITTC_SHT_stock.xml`,
- STIMA skladový a cenový MVP feed `stima-stock-price` z `ITTC_SHT_stock_price.xml`,
- Autronic katalogový MVP feed `autronic-products` filtrovaný na nábytek (`NA-*`) a test feed `autronic-products-test`,
- SEGO katalogový MVP feed `sego` a test feed `sego-test`,
- HON katalogový MVP feed `hon` a test feed `hon-test`,
- endpointy `GET /feeds/*.xml` po ručních rebuild bězích,
- Shoptet XML writer pro jednoduché produkty, varianty, variantní parametry a sklad po skladech,
- endpoint `GET /status` se stavem posledních rebuild běhů,
- konfigurovatelný data adresář přes `DATA_DIR` nebo Railway Volume,
- základní test handleru,
- `Dockerfile` pro Railway deployment,
- `railway.json` s Dockerfile builderem.

Další dodavatelské feedy a plánované spouštění budou doplněné v dalších krocích.

## Požadavky pro lokální vývoj

Stačí Go:

```bash
go version
```

Projekt aktuálně nepoužívá žádné další lokální nástroje typu `air`, `just` nebo Makefile.

## Lokální běh

Nejdřív vygeneruj dummy feed:

```bash
go run ./cmd/rebuild --supplier hello
```

Potom spusť server:

```bash
go run ./cmd/server
```

Server poslouchá na portu z environment variable `PORT`. Pokud není nastavena, použije `8080`.

Feed výstupy a status se ukládají do adresáře `data`. Pro lokální override lze použít `DATA_DIR`:

```bash
DATA_DIR=/tmp/arpari-data go run ./cmd/rebuild --supplier hello
DATA_DIR=/tmp/arpari-data go run ./cmd/rebuild --supplier stima-products
DATA_DIR=/tmp/arpari-data go run ./cmd/rebuild --supplier stima-products-test
DATA_DIR=/tmp/arpari-data go run ./cmd/rebuild --supplier stima-stock
DATA_DIR=/tmp/arpari-data go run ./cmd/rebuild --supplier stima-stock-price
DATA_DIR=/tmp/arpari-data go run ./cmd/rebuild --supplier autronic-products
DATA_DIR=/tmp/arpari-data go run ./cmd/rebuild --supplier autronic-products-test
DATA_DIR=/tmp/arpari-data go run ./cmd/rebuild --supplier sego
DATA_DIR=/tmp/arpari-data go run ./cmd/rebuild --supplier sego-test
DATA_DIR=/tmp/arpari-data go run ./cmd/rebuild --supplier hon
DATA_DIR=/tmp/arpari-data go run ./cmd/rebuild --supplier hon-test
DATA_DIR=/tmp/arpari-data go run ./cmd/server
```

```bash
curl http://localhost:8080/
curl http://localhost:8080/healthz
curl http://localhost:8080/status
curl http://localhost:8080/feeds/hello.xml
curl http://localhost:8080/feeds/stima-products.xml
curl http://localhost:8080/feeds/stima-products-test.xml
curl http://localhost:8080/feeds/stima-stock.xml
curl http://localhost:8080/feeds/stima-stock-price.xml
curl http://localhost:8080/feeds/autronic-products.xml
curl http://localhost:8080/feeds/autronic-products-test.xml
curl http://localhost:8080/feeds/sego.xml
curl http://localhost:8080/feeds/sego-test.xml
curl http://localhost:8080/feeds/hon.xml
curl http://localhost:8080/feeds/hon-test.xml
```

Očekávané odpovědi:

```text
Hello world!
ok
```

Endpoint `/feeds/hello.xml` vrací jednoduchý XML feed s jednou dummy položkou:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<SHOP>
  <SHOPITEM>
    <CODE>HELLO-001</CODE>
    <NAME>Hello world product</NAME>
    <PRICE_VAT>123.45</PRICE_VAT>
    <STOCK>7</STOCK>
  </SHOPITEM>
</SHOP>
```

Endpoint `/feeds/stima-products.xml` vrací technické MVP STIMA katalogu. Obsahuje základní produktová/variantní pole, krátký i dlouhý popis, sklad, cílové Shoptet kategorie, obrázky a povolené variantní parametry `KOSTRA`, `Sedák`, `Délka stolu`, `Rozklad`. Produkty s více než 512 variantami se při transformaci ořežou na prvních 512 variant v pořadí ze zdrojového feedu. Položky bez bezpečně určené cílové kategorie se přeskočí.

Endpoint `/feeds/stima-products-test.xml` vrací stejnou katalogovou transformaci jako `stima-products`, ale jen prvních 20 výstupních produktů. Slouží pro rychlé ruční testy v Shoptetu.

Endpointy `/feeds/stima-stock.xml` a `/feeds/stima-stock-price.xml` vrací aktualizační MVP pro STIMA sklad, respektive sklad + cenu. Neobsahují katalogová pole jako popisy, obrázky nebo kategorie. I tyto feedy respektují Shoptet limit 512 variant na produkt.

Endpoint `/feeds/autronic-products.xml` vrací katalogový MVP Autronicu filtrovaný na nábytkové kategorie s prefixem `NA-`. Obsahuje kód, název, EAN, cenu s DPH, sklad, sklad po skladech, popis, obrázky a základní mapování do cílových Shoptet kategorií. Endpoint `/feeds/autronic-products-test.xml` vrací prvních 20 výstupních produktů pro rychlý ruční import.

Endpoint `/feeds/sego.xml` vrací katalogový MVP ze SEGO Zboží.cz styl feedu. Obsahuje kód, název, EAN, cenu s DPH, dostupnost, popis, obrázky a základní mapování do kancelářských židlí. Endpoint `/feeds/sego-test.xml` vrací prvních 20 produktů.

Endpoint `/feeds/hon.xml` vrací katalogový MVP z HON feedu. Obsahuje kód, název, cenu s DPH, sklad, dostupnost, popis, obrázky a základní mapování do kancelářských židlí / bytových doplňků. Endpoint `/feeds/hon-test.xml` vrací prvních 20 produktů.

Endpoint `/status` vrací stav posledních lokálních rebuild běhů:

```json
{
  "feeds": {
    "hello": {
      "filename": "hello.xml",
      "lastRunAt": "2026-05-24T12:34:56Z",
      "status": "success",
      "itemsProcessed": 1,
      "itemsSkipped": 0,
      "error": ""
    }
  }
}
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
- Railway Volume mount path z environment variable `RAILWAY_VOLUME_MOUNT_PATH`, pokud je ke službě připojený Volume.

Aplikace v kontejneru poslouchá na:

```text
0.0.0.0:${PORT}
```

Pro persistentní feedy na Railway:

1. Připoj ke službě Railway Volume.
2. Nastav mount path Volume na `/data`.
3. Nenastavuj ručně `DATA_DIR`; aplikace použije `RAILWAY_VOLUME_MOUNT_PATH`.
4. Po deployi ověř `/status`; `/feeds/hello.xml` ověř po prvním rebuild běhu, který feed do Volume uloží.

Priority data adresáře jsou:

```text
DATA_DIR -> RAILWAY_VOLUME_MOUNT_PATH -> data
```

Po deployi ověř:

```text
https://<railway-domain>/
https://<railway-domain>/healthz
https://<railway-domain>/status
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

U Autronicu platí pevné zadání: brát pouze kategorii nábytek. Všechny ostatní kategorie musí parser přeskočit.

Druhá fáze:

- Sakypaky: `https://www.sakypaky.cz/export/b2b_partners_cs.xml`,
- Dřevočal: `https://www.matrace-drevocal.cz/feed/`.

Poznámka k Autronic dostupnostnímu feedu: `HEAD` request vrací 404, ale běžný `GET` vrací XML. Parser proto musí používat `GET`.

## Dokumentace Shoptet

- [Shoptet XML specifikace](https://developers.shoptet.com/shoptet-tools/shoptet-xml-specification/)
- [Automatické importy produktů](https://podpora.shoptet.cz/automaticke-importy-produktu/)
- [Šablony variant](https://podpora.shoptet.cz/sablony-variant/)
- [Příplatkové parametry](https://podpora.shoptet.cz/priplatkove-parametry/)
- [Shoptet XML validátor](https://www.shoptet.cz/xml-validace/) pro ruční externí kontrolu feed URL.
