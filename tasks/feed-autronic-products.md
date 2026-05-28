# Feed Task: Autronic katalog produktů

## Metadata

- Supplier: Autronic
- Generator name: `autronic-products`
- Output endpoint: `/feeds/autronic-products.xml`
- Test generator name: `autronic-products-test`
- Test output endpoint: `/feeds/autronic-products-test.xml`
- Source URL: `https://autronic.cz/feeds/product-feed.xml`
- Priority: první fáze
- Status: MVP implementováno, čeká na ruční importní ladění
- Last updated: 2026-05-28

## Cíl

Generovat Shoptet produktový XML feed z Autronic katalogu pro samostatný testovací import.

## Aktuální pravidla

- Reálný aktuální Autronic katalog má 5 750 položek.
- Povinně importovat pouze kategorii nábytek.
- MVP filtr bere produkty s `CategoryShortName` prefixem `NA-`.
- Reálný výstup po filtru má 800 produktů.
- Všechny ostatní kategorie musí parser vyřadit před generováním Shoptet XML.
- Testovací endpoint `autronic-products-test` používá stejná pravidla, ale končí po prvních 5 výstupních produktech.
- Katalog nesmí bez mapování přepsat citlivá produktová data původního katalogu.

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Základní katalogová pole podle bezpečného zadání:
  - `NAME`
  - `EAN`
  - `PRICE_VAT`
  - `STOCK`
  - sklad po skladech
  - `DESCRIPTION`
  - `IMAGES`
  - základní cílové kategorie podle názvu zdrojové kategorie

## Implementace

- Stav kódu: implementováno v `internal/autronic/products.go` a `internal/feed/autronic_products.go`.
- Registry: supplier `autronic-products` a `autronic-products-test` jsou dostupné přes `cmd/rebuild`.
- Lokální testy: `go test ./...` prochází.
- Reálný rebuild ověřen 2026-05-28:
  - `autronic-products`: 5 750 přečteno, 800 emitováno, 4 950 přeskočeno.
  - `autronic-products-test`: 5 emitovaných produktů.
  - Výstupní XML je well-formed a publikace proběhla přes storage publisher.

## Otevřené otázky

- Potvrdit, zda prefix `NA-` přesně odpovídá klientem chtěné kategorii nábytek.
- Potvrdit, zda do první vlny patří i bytové doplňky `BD-*`.
- Která pole smí Autronic katalog zakládat nebo přepisovat?
- Doladit cílové kategorie po ručním importním testu.

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier autronic-products` vytvoří validní XML.
- Výstup je dostupný na `/feeds/autronic-products.xml`.
- `go run ./cmd/rebuild --supplier autronic-products-test` vytvoří testovací feed s 5 produkty.
- Testovací výstup je dostupný na `/feeds/autronic-products-test.xml`.
- Do výstupu se dostanou pouze produkty z kategorie nábytek.
- Výstup nepřekročí Shoptet limit položek bez dalšího rozdělení.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test položky z kategorie nábytek.
- Unit test přeskočení jiné kategorie.
- Unit test limitu počtu položek.
- Unit test chybějícího `CODE`.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
