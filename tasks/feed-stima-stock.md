# Feed Task: STIMA sklad

## Metadata

- Supplier: STIMA
- Generator name: `stima-stock`
- Output endpoint: `/feeds/stima-stock.xml`
- Source URL: `https://www.stima.cz/userfiles/xml/ITTC_SHT_stock.xml`
- Priority: první fáze
- Status: draft, čeká na implementaci

## Cíl

Generovat samostatný Shoptet XML feed pro aktualizaci skladů STIMA produktů a variant.

## Aktuální pravidla

- Feed má aktualizovat jen skladová data.
- Nemá přepisovat katalogová pole jako názvy, popisy, obrázky, SEO nebo kategorie.
- Musí používat `GET` download a standardní bezpečný publish flow.
- Chyba jednoho skladového běhu nesmí rozbít katalogový feed.

## MVP rozsah

- Identifikace produktu nebo varianty:
  - `CODE`
- Sklad:
  - `STOCK`
  - případně sklad po skladech, pokud ho zdroj obsahuje

## Otevřené otázky

- Má Shoptet import skladů očekávat jednoduchý `STOCK`, nebo strukturu `WAREHOUSES`?
- Budeme sklad u STIMA variant řešit podle variant `CODE`, nebo přes mapování na původní katalog?
- Jak často sklad aktualizovat po nasazení automatického cronu?
- Má skladový feed pomoci rozlišit `Skladem / Na zakázku` u STIMA katalogových variant?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier stima-stock` vytvoří validní XML.
- Výstup je dostupný na `/feeds/stima-stock.xml`.
- Výstup obsahuje pouze bezpečná skladová pole.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test běžné skladové položky.
- Unit test položky bez `CODE`.
- Unit test prázdného nebo nevalidního zdroje.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
