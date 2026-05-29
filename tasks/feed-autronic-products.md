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
- Last updated: 2026-05-29

## Cíl

Generovat Shoptet produktový XML feed z Autronic katalogu pro samostatný testovací import.

## Aktuální pravidla

- Reálný aktuální Autronic katalog má 5 744 hlavních produktů.
- Importuje se celý Autronic nábytek, tedy všechny produkty s `CategoryShortName` prefixem `NA-`.
- Kategorie sezóna/dekorace se neimportují; produkty s `CategoryShortName` prefixem `DE-` se vyřazují.
- Z bytových doplňků `BD-*` se importují pouze klientem schválené podkategorie: věšáky, poličky, regály, organizéry, odpadkové koše, stojany na šaty, botníky, paravány, taburety a němí sluhové.
- Z bytových doplňků se vyřazují například stolování, hodiny a obrazy, koše prádelní a regálové, ubrusy a polštáře, zrcadla, stojany na květiny, nástěnné dekorace, fotorámečky a stojany na dřevo.
- Reálný výstup po filtru má 1 093 produktů.
- Testovací endpoint `autronic-products-test` používá stejná pravidla, ale končí po prvních 5 výstupních produktech.
- Katalog nesmí bez mapování přepsat citlivá produktová data původního katalogu.
- Každý produkt se označuje `SUPPLIER=Autronic`; hodnota je určená pro interní filtrování dodavatele v Shoptet administraci.
- Zdrojové `Parameters/Parameter` hodnoty se exportují jako Shoptet `INFORMATION_PARAMETERS`, tedy jako tabulkové doplňkové parametry, ne jako vybíratelné varianty.

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
  - technické parametry v `INFORMATION_PARAMETERS`
  - cílové kategorie podle zdrojového `CategoryShortName`

## Implementace

- Stav kódu: implementováno v `internal/autronic/products.go` a `internal/feed/autronic_products.go`.
- Registry: supplier `autronic-products` a `autronic-products-test` jsou dostupné přes `cmd/rebuild`.
- Lokální testy: `go test ./...` prochází.
- Reálný rebuild ověřen 2026-05-29:
  - `autronic-products`: 5 744 přečteno, 1 093 emitováno, 4 651 přeskočeno.
  - `autronic-products-test`: 5 emitovaných produktů.
  - `autronic-products`: 1 093 bloků `INFORMATION_PARAMETERS`.
  - Výstupní XML je well-formed a publikace proběhla přes storage publisher.

## Otevřené otázky

- Která pole smí Autronic katalog zakládat nebo přepisovat?
- Doladit cílové kategorie po ručním importním testu.

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier autronic-products` vytvoří validní XML.
- Výstup je dostupný na `/feeds/autronic-products.xml`.
- `go run ./cmd/rebuild --supplier autronic-products-test` vytvoří testovací feed s 5 produkty.
- Testovací výstup je dostupný na `/feeds/autronic-products-test.xml`.
- Do výstupu se dostane celý Autronic nábytek `NA-*` a pouze schválené bytové doplňky `BD-*`.
- Sezóna/dekorace `DE-*` se do výstupu nedostane.
- Výstup nepřekročí Shoptet limit položek bez dalšího rozdělení.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test položky z kategorie nábytek.
- Unit test přeskočení jiné kategorie.
- Unit test pravidel pro schválené a vyřazené Autronic kategorie.
- Unit test limitu počtu položek.
- Unit test chybějícího `CODE`.
- Rebuild test přes fixture-backed downloader.
- Ruční kontrola přes Shoptet XML validátor.
