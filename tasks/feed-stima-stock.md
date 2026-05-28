# Feed Task: STIMA sklad

## Metadata

- Supplier: STIMA
- Generator name: `stima-stock`
- Output endpoint: `/feeds/stima-stock.xml`
- Source URL: `https://www.stima.cz/userfiles/xml/ITTC_SHT_stock.xml`
- Priority: první fáze
- Status: MVP implementováno, business pravidla k doplnění
- Last updated: 2026-05-28

## Cíl

Generovat samostatný Shoptet XML feed pro aktualizaci skladů STIMA produktů a variant.

## Aktuální pravidla

- Feed má aktualizovat jen skladová data.
- Nemá přepisovat katalogová pole jako názvy, popisy, obrázky, SEO nebo kategorie.
- Musí používat `GET` download a standardní bezpečný publish flow.
- Chyba jednoho skladového běhu nesmí rozbít katalogový feed.
- Produkty nad 512 variant se zatím oříznou na prvních 512 variant v pořadí ze STIMA feedu.
- U variantních produktů se parent kód zapisuje jako Shoptet `EXTERNAL_ID`, ne jako top-level `CODE`; jednotlivé varianty dál nesou vlastní `CODE`.

## MVP rozsah

- Identifikace produktu nebo varianty:
  - `CODE`
  - parent `EXTERNAL_ID` u variantních produktů
- Sklad:
  - `STOCK`
  - případně sklad po skladech, pokud ho zdroj obsahuje

## Aktuální ověření

- Stav kódu: implementováno v `internal/stima/updates.go` a `internal/feed/stima_updates.go`.
- Registry: supplier `stima-stock` je dostupný přes `cmd/rebuild`.
- Lokální testy: `go test ./...` prochází.
- Reálný rebuild proti STIMA zdroji ověřen 2026-05-28.
- Reálný výstup ověřen proti Shoptet RNG schématu `products-supplier-v10.rng`.
- Poslední ověřené počty:
  - products read: 953
  - products emitted: 953
  - products skipped: 0
  - products trimmed over 512 variants: 20
  - variants emitted: 20581
  - variants trimmed: 3431
  - output size: přibližně 11.6 MB

## Otevřené otázky

- Má Shoptet import skladů očekávat jednoduchý `STOCK`, nebo strukturu `WAREHOUSES`?
- Budeme sklad u STIMA variant řešit podle variant `CODE`, nebo přes mapování na původní katalog?
- Jak často sklad aktualizovat po nasazení automatického cronu?
- Má skladový feed pomoci rozlišit `Skladem / Na zakázku` u STIMA katalogových variant?
- Je pro skladové aktualizace lepší zachovat variantní strukturu, nebo posílat varianty po jednotlivých kódech?

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
