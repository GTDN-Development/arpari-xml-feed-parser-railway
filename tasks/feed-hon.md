# Feed Task: HON katalog

## Metadata

- Supplier: HON
- Generator name: `hon`
- Output endpoint: `/feeds/hon.xml`
- Source URL: `https://www.webshop.officepro-brno.cz/import/HONClientFeed/HONClientFeed.xml`
- Priority: první fáze
- Status: draft, čeká na implementaci

## Cíl

Generovat Shoptet XML feed z HON dodavatelského katalogu.

## Aktuální pravidla

- HON má vlastní XML strukturu.
- Parser musí vycházet z reálné struktury feedu, ne z předpokladu Zboží.cz / Heureka.
- Katalogová data nesmí bez mapování přepisovat citlivá data původního katalogu.
- Kategorie bude potřeba řešit přes společné mapování.

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Bezpečná základní pole:
  - `NAME`
  - `EAN`, pokud je ve zdroji
  - `PRICE_VAT`, pokud je ve zdroji
  - `STOCK` nebo dostupnost, pokud je ve zdroji

## Otevřené otázky

- Jaká je přesná XML struktura HON feedu?
- Které pole je stabilní produktový kód?
- Obsahuje feed varianty nebo konfigurace?
- Jak mapovat HON kategorie na Shoptet?
- Má HON zakládat nové produkty, nebo jen aktualizovat existující?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier hon` vytvoří validní XML.
- Výstup je dostupný na `/feeds/hon.xml`.
- Výstup nepřekročí Shoptet limity.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test běžné HON položky.
- Unit test chybějícího produktu kódu.
- Unit test variant, pokud je zdroj obsahuje.
- Unit test kategorie po doplnění mapování.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
