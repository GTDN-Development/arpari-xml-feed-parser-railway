# Feed Task: SEGO katalog

## Metadata

- Supplier: SEGO
- Generator name: `sego`
- Output endpoint: `/feeds/sego.xml`
- Test generator name: `sego-test`
- Test output endpoint: `/feeds/sego-test.xml`
- Source URL: `https://segocz.cz/src/Frontend/Files/Feeds/Catalog/zbozi_123456.xml`
- Priority: první fáze
- Status: MVP implementováno, čeká na ruční importní ladění
- Last updated: 2026-05-28

## Cíl

Generovat Shoptet XML feed ze SEGO katalogového feedu.

## Aktuální pravidla

- Zdroj je ve stylu Zboží.cz.
- Transformace do Shoptet XML struktury je implementovaná jako katalogový MVP.
- Reálný zdroj má 144 položek.
- Testovací endpoint `sego-test` používá stejná pravidla, ale končí po prvních 5 produktech.
- Katalogová data nesmí bez mapování přepisovat citlivá data původního katalogu.
- Kategorie se mapují na cílové Shoptet kategorie včetně podkategorií podle názvu, popisu a parametrů SEGO položky.
- Normalizovaný export Shoptet kategorií je uložený v `reference/shoptet-categories.csv`.
- SEGO flat varianty typu `Produkt | Hodnota` se slučují do Shoptet variant podle produktového URL slug a odpovídajícího zdrojového parametru.
- SEGO `Catalog/VariantImages/.../previewImg...` URL se do výstupu neposílají; zvenku vrací 404 a Shoptet je při importu nestáhne. Pro `IMAGES` a variantní `IMAGE_REF` se používají funkční `Catalog/.../source/...` URL.
- Variantní produkty používají tvar porovnaný s exportem ručně nastaveného Shoptet produktu: parent nemá `CODE` ani `EXTERNAL_ID`, varianty nesou vlastní `CODE`, `CURRENCY`, `VAT`, `PRICE_VAT`, `AVAILABILITY`, `IMAGE_REF` a `PARAMETERS`.
- SEGO obrázky jsou omezené na prvních 20 funkčních URL na produkt, aby import neposílal desítky duplicitních nebo doplňkových fotek na jeden variantní parent.
- Variantní parametr se obecně bere ze zdrojového `PARAM_NAME`; nepřejmenováváme hodnoty heuristicky, pokud to není pro SEGO nutné. Aktuální výjimka: rozměrové hodnoty typu `150x220mm`, které zdroj posílá jako `Barva`, se exportují jako `Rozměr`. Aby se na detailu zobrazily kulaté vzorníky jako v referenčním e-shopu, musí v administraci/šabloně variant existovat odpovídající parametr a všechny použité hodnoty musí mít nastavenou barvu nebo obrázek; XML feed nastavuje hodnotu varianty a `IMAGE_REF`, ne vizuál vzorníku hodnoty.
- SEGO ceny se exportují jako celé Kč v `PRICE_VAT` s `VAT=21` a `CURRENCY=CZK`; desetinné ceny ze zdroje se zaokrouhlují.

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Bezpečná základní pole:
  - `NAME`
  - `EAN`, pokud je ve zdroji
  - `PRICE_VAT`, `VAT` a `CURRENCY`, pokud je ve zdroji cena
  - dostupnost z `DELIVERY_DATE`
  - `DESCRIPTION`
  - `IMAGES`
- Cílové kategorie:
  - síťované kancelářské židle
  - čalouněná kancelářská křesla
  - dětské kancelářské židle
  - pracovní a průmyslové židle
  - náhradní díly a podložky
  - konferenční židle
  - laboratorní židle
  - zdravotní židle

## Implementace

- Stav kódu: implementováno v `internal/sego/products.go`, `internal/sego/categories.go` a `internal/feed/sego_products.go`.
- Registry: supplier `sego` a `sego-test` jsou dostupné přes `cmd/rebuild`.
- Lokální testy: `go test ./...` prochází.
- Reálný rebuild ověřen 2026-05-28:
  - `sego`: 144 přečteno, 100 emitovaných Shoptet produktů.
  - `sego`: 37 produktů s variantami, 81 emitovaných variant.
  - `sego-test`: 5 emitovaných Shoptet produktů.
  - Výstupní XML je well-formed a publikace proběhla přes storage publisher.
- Externí Shoptet validace je samostatný povinný krok a může odhalit chyby, které lokální well-formed kontrola nevidí.
- Oficiální validátor: https://www.shoptet.cz/xml-validace/

## Otevřené otázky

- Potvrdit, jestli `ITEM_ID` je správný stabilní produktový kód pro párování v Shoptetu.
- Potvrdit, zda SEGO produkty mají zůstat jako jednoduché produkty, nebo se mají později slučovat do variant.
- Doladit cílové kategorie po ručním importním testu.
- Má SEGO zakládat nové produkty, nebo jen aktualizovat existující?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier sego` vytvoří validní XML.
- Výstup je dostupný na `/feeds/sego.xml`.
- `go run ./cmd/rebuild --supplier sego-test` vytvoří testovací feed s 5 produkty.
- Testovací výstup je dostupný na `/feeds/sego-test.xml`.
- Výstup nepřekročí Shoptet limity.
- Veřejná URL `/feeds/sego.xml` projde ruční kontrolou v Shoptet XML validátoru proti produktové dodavatelské Relax NG specifikaci.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test běžné SEGO položky.
- Unit test chybějícího produktu kódu.
- Unit test limitu testovacího výstupu.
- Unit test základního mapování kategorie.
- Unit test mapování SEGO podkategorií.
- Unit test slučování flat barevných variant.
- Unit test feed-specific opravy SEGO rozměru chybně poslaného jako `Barva`.
- Unit test filtrování nefunkčních SEGO variant preview obrázků.
- Rebuild test přes fixture-backed downloader.
- Po každé změně SEGO transformace spustit ruční kontrolu veřejné URL přes Shoptet XML validátor: https://www.shoptet.cz/xml-validace/
