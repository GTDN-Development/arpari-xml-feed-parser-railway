# Feed Task: SEGO katalog

## Metadata

- Supplier: SEGO
- Generator name: `sego`
- Output endpoint: `/feeds/sego.xml`
- Test generator name: `sego-test`
- Test output endpoint: `/feeds/sego-test.xml`
- Source URL: `https://segocz.cz/src/Frontend/Files/Feeds/Catalog/heureka_feed.xml`
- Priority: první fáze
- Status: MVP implementováno, čeká na ruční importní ladění
- Last updated: 2026-06-04

## Cíl

Generovat Shoptet XML feed ze SEGO katalogového feedu.

## Rozhodnutí

- 2026-06-04: Produkční endpoint `/feeds/sego.xml` zůstává beze změny, aby se nemusel měnit automatický import v Shoptetu.
- 2026-06-04: Zdroj SEGO se mění ze `zbozi_123456.xml` na `heureka_feed.xml`, protože obsahuje stejná základní katalogová data a navíc `VIDEO_URL` u části položek.
- 2026-06-04: Video zatím nepřenášíme do výstupního Shoptet XML. Podpora videa je navazující krok, protože Shoptet video elementy patří do complete XML schématu (`RELATED_VIDEOS` / `RELATED_VIDEO` / `YOUTUBE_VIDEO_CODE`) a musí se samostatně ověřit proti konkrétnímu typu automatického importu.
- 2026-06-09: SEGO výstup používá jako Shoptet `CODE` primárně `EAN`, ne Heureka `ITEM_ID`. Starý zdroj `zbozi_123456.xml` měl `ITEM_ID == EAN`, zatímco nový `heureka_feed.xml` posílá `ITEM_ID` s prefixem. Párování podle nového `ITEM_ID` by v Shoptetu zakládalo duplicity.
- 2026-06-09: Nový Heureka zdroj posílá u některých boolean parametrů interní labely `{$lblCoreYesLabel}` / `{$lblCoreNoLabel}`. Transformace je normalizuje na `Ano` / `Ne`, aby se šablonové tokeny nepropsaly na detail produktu.
- 2026-06-24: SEGO feed vynechává produkty uvedené v `internal/sego/excluded_products.csv`, protože je klient spravuje ručně v Shoptetu a automatický import by u nich zakládal duplicity.

## Aktuální pravidla

- Zdroj je Heureka feed ve struktuře blízké původnímu Zboží.cz feedu.
- Transformace do Shoptet XML struktury je implementovaná jako katalogový MVP.
- Reálný zdroj má 152 položek.
- Testovací endpoint `sego-test` používá stejná pravidla, ale končí po prvních 5 produktech.
- Katalogová data nesmí bez mapování přepisovat citlivá data původního katalogu.
- Každý produkt se označuje `SUPPLIER=SEGO`; hodnota je určená pro interní filtrování dodavatele v Shoptet administraci.
- Kategorie se mapují na cílové Shoptet kategorie včetně podkategorií podle názvu, popisu a parametrů SEGO položky.
- Normalizovaný export Shoptet kategorií je uložený v `reference/shoptet-categories.csv`.
- SEGO flat varianty typu `Produkt | Hodnota` se slučují do Shoptet variant podle produktového URL slug a odpovídajícího zdrojového parametru.
- Identifikace produktů a variant používá stabilní EAN, pokud je ve zdroji dostupný; `ITEM_ID` je pouze fallback pro položky bez EAN.
- Produkty z blocklistu se vynechávají podle stejného stabilního kódu před skládáním variant. Aktuální blocklist obsahuje 50 EANů.
- SEGO `Catalog/VariantImages/.../previewImg...` URL se do výstupu neposílají; zvenku vrací 404 a Shoptet je při importu nestáhne. Pro `IMAGES` a variantní `IMAGE_REF` se používají funkční `Catalog/.../source/...` URL.
- Variantní produkty používají tvar porovnaný s exportem ručně nastaveného Shoptet produktu: parent nemá `CODE` ani `EXTERNAL_ID`, varianty nesou vlastní `CODE`, `CURRENCY`, `VAT`, `PRICE_VAT`, `AVAILABILITY`, `IMAGE_REF` a `PARAMETERS`.
- SEGO obrázky jsou omezené na prvních 20 funkčních URL na produkt, aby import neposílal desítky duplicitních nebo doplňkových fotek na jeden variantní parent.
- Variantní parametr se obecně bere ze zdrojového `PARAM_NAME`; nepřejmenováváme hodnoty heuristicky, pokud to není pro SEGO nutné. Aktuální výjimka: rozměrové hodnoty typu `150x220mm`, které zdroj posílá jako `Barva`, se exportují jako `Rozměr`. Aby se na detailu zobrazily kulaté vzorníky jako v referenčním e-shopu, musí v administraci/šabloně variant existovat odpovídající parametr a všechny použité hodnoty musí mít nastavenou barvu nebo obrázek; XML feed nastavuje hodnotu varianty a `IMAGE_REF`, ne vizuál vzorníku hodnoty.
- SEGO ceny se exportují jako celé Kč v `PRICE_VAT` s `VAT=21` a `CURRENCY=CZK`; desetinné ceny ze zdroje se zaokrouhlují.
- Technické `PARAM` hodnoty ze zdroje se exportují včetně jednotek (`UNIT`) jako Shoptet `INFORMATION_PARAMETERS`, aby se zobrazily v tabulce doplňkových parametrů. Parametr použitý jako volba varianty se na parent produktu neduplikuje jako informační parametr.
- Boolean labely `{$lblCoreYesLabel}` a `{$lblCoreNoLabel}` se ve výstupu převádějí na `Ano` a `Ne`.

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
  - technické parametry v `INFORMATION_PARAMETERS`
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

- Stav kódu: implementováno v `internal/sego/products.go`, `internal/sego/categories.go`, `internal/sego/exclusions.go` a `internal/feed/sego_products.go`.
- Verzovaný blocklist produktů je v `internal/sego/excluded_products.csv`.
- Registry: supplier `sego` a `sego-test` jsou dostupné přes `cmd/rebuild`.
- Lokální testy: `go test ./...` prochází.
- Reálný rebuild ověřen 2026-06-26:
  - `sego`: 152 přečteno, 50 vynecháno přes blocklist, 74 emitovaných Shoptet produktů.
  - `sego`: 16 produktů s variantami, 44 emitovaných variant.
  - `sego-test`: 5 emitovaných Shoptet produktů, 50 vynecháno přes blocklist.
  - Výstupní XML je well-formed a publikace proběhla přes storage publisher.
- Externí Shoptet validace je samostatný povinný krok a může odhalit chyby, které lokální well-formed kontrola nevidí.
- Oficiální validátor: https://www.shoptet.cz/xml-validace/

## Otevřené otázky

- Potvrdit, zda SEGO produkty mají zůstat jako jednoduché produkty, nebo se mají později slučovat do variant.
- Doladit cílové kategorie po ručním importním testu.
- Má SEGO zakládat nové produkty, nebo jen aktualizovat existující?

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier sego` vytvoří validní XML.
- Výstup je dostupný na `/feeds/sego.xml`.
- `go run ./cmd/rebuild --supplier sego-test` vytvoří testovací feed s 5 produkty.
- Testovací výstup je dostupný na `/feeds/sego-test.xml`.
- Výstup neobsahuje žádný kód uvedený v `internal/sego/excluded_products.csv`.
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
