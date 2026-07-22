# Zpracování dodavatelských feedů

Parser je mezivrstva mezi dodavatelskými XML feedy a Shoptetem. Nestahuje data beze změny, ale převádí je do formátu, který Shoptet umí bezpečně importovat.

Produkční adresa parseru je:

```text
https://arpari-xml-feed-parser-railway-production.up.railway.app/
```

## Kdy se spouští kompletní rebuild

Všechny ostré feedy se automaticky znovu generují každý den ve `02:00 UTC`. To znamená, že se stáhnou aktuální data od všech dodavatelů a znovu se vytvoří všechny veřejné XML výstupy.

V českém čase to znamená:

- `04:00` během letního času
- `03:00` během zimního času

Při každém běhu parser stáhne aktuální dodavatelské XML, převede ho a teprve po úspěšné kontrole zveřejní nový výstup. Pokud je problém ve zdrojovém feedu nebo transformaci, poslední funkční XML zůstane zachované.

## Obecně

- Vybereme jen pole, která jsou pro daný feed povolená.
- Položky bez důležité identifikace, například bez kódu nebo názvu, vynecháme.
- Kategorie dodavatele převádíme na cílové kategorie v Shoptetu. Pokud kategorii neumíme bezpečně určit, produkt raději nepošleme.
- Varianty skládáme do variantního modelu Shoptetu. Shoptet má limit 512 variant na jeden produkt.
- Katalogové feedy mohou posílat názvy, popisy, obrázky, kategorie a parametry.
- Aktualizační feedy posílají jen sklad, případně ceny, aby nepřepisovaly katalogová data.

Výstupní Shoptet XML má vždy základní strukturu `SHOP` -> `SHOPITEM`. Jednoduché produkty mají vlastní `CODE`. Variantní produkty mají parent položku a varianty uvnitř `VARIANTS` -> `VARIANT`; kódy jsou potom hlavně na úrovni jednotlivých variant.

## Feedy

### STIMA katalog

Adresa vstupu od dodavatele:

```text
https://www.stima.cz/userfiles/xml/ITTC_SHT_products.xml
```

Stažení vstupu od dodavatele:

```bash
curl -L -o source-stima-products.xml https://www.stima.cz/userfiles/xml/ITTC_SHT_products.xml
```

Adresa výstupu parseru:

```text
https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/stima-products.xml
```

Stažení výstupu parseru:

```bash
curl -L -o parser-stima-products.xml https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/stima-products.xml
```

Co děláme:

- Produkt identifikujeme podle `CODE`; u variant bereme kódy z `VARIANTS` -> `VARIANT` -> `CODE`.
- Do výstupu posíláme název, krátký a dlouhý popis, EAN, cenu, sklad, obrázky a doplňkové parametry.
- Kategorie bereme ze vstupních `CATEGORIES` -> `CATEGORY` a mapujeme je hlavně do židlí a stolů v Shoptetu.
- Variantní parametry bereme z `VARIANTS` -> `VARIANT` -> `PARAMETERS` -> `PARAMETER`; pouštíme jen `KOSTRA`, `Sedák`, `Délka stolu`, `Rozklad`, `Specifikace`.
- Produkty bez bezpečné kategorie nebo nedostupné produkty bez obrázku vynecháme.

### STIMA sklad

Adresa vstupu od dodavatele:

```text
https://www.stima.cz/userfiles/xml/ITTC_SHT_stock.xml
```

Stažení vstupu od dodavatele:

```bash
curl -L -o source-stima-stock.xml https://www.stima.cz/userfiles/xml/ITTC_SHT_stock.xml
```

Adresa výstupu parseru:

```text
https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/stima-stock.xml
```

Stažení výstupu parseru:

```bash
curl -L -o parser-stima-stock.xml https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/stima-stock.xml
```

Co děláme:

- Posíláme jen identifikaci produktu nebo varianty a sklad.
- Pokud zdroj obsahuje sklady po skladech, přenášíme `STOCK` -> `WAREHOUSES` -> `WAREHOUSE`.
- U variant držíme tvar `VARIANTS` -> `VARIANT`, aby Shoptet správně poznal variantní produkty.
- Neposíláme názvy, popisy, obrázky, kategorie ani ceny.

### STIMA sklad a ceny

Adresa vstupu od dodavatele:

```text
https://www.stima.cz/userfiles/xml/ITTC_SHT_stock_price.xml
```

Stažení vstupu od dodavatele:

```bash
curl -L -o source-stima-stock-price.xml https://www.stima.cz/userfiles/xml/ITTC_SHT_stock_price.xml
```

Adresa výstupu parseru:

```text
https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/stima-stock-price.xml
```

Stažení výstupu parseru:

```bash
curl -L -o parser-stima-stock-price.xml https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/stima-stock-price.xml
```

Co děláme:

- Posíláme sklad stejně jako ve skladovém feedu.
- Navíc posíláme cenu s DPH.
- Identifikace produktu nebo varianty zůstává přes `CODE`; u variant přes `VARIANTS` -> `VARIANT` -> `CODE`.
- Nepřepisujeme katalogová pole jako názvy, popisy, obrázky nebo kategorie.

### Autronic katalog

Adresa vstupu od dodavatele:

```text
https://autronic.cz/feeds/product-feed.xml
```

Stažení vstupu od dodavatele:

```bash
curl -L -o source-autronic-products.xml https://autronic.cz/feeds/product-feed.xml
```

Adresa výstupu parseru:

```text
https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/autronic-products.xml
```

Stažení výstupu parseru:

```bash
curl -L -o parser-autronic-products.xml https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/autronic-products.xml
```

Co děláme:

- Do výstupu pouštíme nábytek a schválené bytové doplňky podle `ProductCategory` -> `CategoryShortName`.
- Dekorace, sezónní sortiment a další neschválené kategorie vynecháváme.
- Produkt identifikujeme podle `ProductCode`.
- Cenu bereme primárně z `Prices` -> `RetailPromotionalPriceIncludingVat`, jinak z `RetailPriceIncludingVat`.
- Sklad bereme z `Availability` -> `StockAvailabilityTotal` a sklady po skladech z `StockAvailability` -> `Stock`.
- Popis, obrázky a parametry posíláme ze zdrojových bloků `Descriptions`, `Images` a `Parameters`.
- Barevné varianty skládáme podle `ColorVariants` -> `Product`; variantní parametr v Shoptetu je `Barva`.

### Autronic sklad

Adresy vstupů od dodavatele:

```text
https://autronic.cz/feeds/availability-feed.xml
https://autronic.cz/feeds/product-feed.xml
```

Stažení vstupů od dodavatele:

```bash
curl -L -o source-autronic-availability.xml https://autronic.cz/feeds/availability-feed.xml
curl -L -o source-autronic-products.xml https://autronic.cz/feeds/product-feed.xml
```

Adresa výstupu parseru:

```text
https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/autronic-availability.xml
```

Stažení výstupu parseru:

```bash
curl -L -o parser-autronic-availability.xml https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/autronic-availability.xml
```

Co děláme:

- Slouží jen pro aktualizaci skladu.
- Skladový feed identifikuje produkty přes `ProductCode`.
- Sklad bereme z `Availability` -> `StockAvailabilityTotal`; sklady po skladech z `Availability` -> `StockAvailability` -> `Stock`.
- Skladový feed filtrujeme podle Autronic katalogu, aby se neposílaly produkty mimo náš sortiment.
- Neposíláme názvy, popisy, ceny, obrázky ani kategorie.

### SEGO katalog

Adresa vstupu od dodavatele:

```text
https://segocz.cz/src/Frontend/Files/Feeds/Catalog/heureka_feed.xml
```

Stažení vstupu od dodavatele:

```bash
curl -L -o source-sego.xml https://segocz.cz/src/Frontend/Files/Feeds/Catalog/heureka_feed.xml
```

Adresa výstupu parseru:

```text
https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/sego.xml
```

Stažení výstupu parseru:

```bash
curl -L -o parser-sego.xml https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/sego.xml
```

Co děláme:

- Jako hlavní kód používáme `EAN`; pokud chybí, použijeme `ITEM_ID`.
- Dostupnost odvozujeme z `DELIVERY_DATE`: `0` převádíme na `Skladem`, zápornou hodnotu na `Momentálně nedostupné` a ostatní hodnoty na počet dnů dodání.
- Obrázky bereme z hlavního obrázku a alternativních obrázků (`IMGURL`, `IMGURL_ALTERNATIVE`).
- Doplňkové parametry bereme z `PARAM`, včetně případné jednotky `UNIT`.
- Produkty ve tvaru `PRODUCTNAME` jako `Název | Hodnota` skládáme do variant, pokud hodnotu najdeme i mezi parametry (`PARAM`).
- Předem domluvené ručně spravované produkty do výstupu neposíláme.

### HON katalog

Adresa vstupu od dodavatele:

```text
https://www.webshop.officepro-brno.cz/import/HONClientFeed/HONClientFeed.xml
```

Stažení vstupu od dodavatele:

```bash
curl -L -o source-hon.xml https://www.webshop.officepro-brno.cz/import/HONClientFeed/HONClientFeed.xml
```

Adresa výstupu parseru:

```text
https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/hon.xml
```

Stažení výstupu parseru:

```bash
curl -L -o parser-hon.xml https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/hon.xml
```

Co děláme:

- Jako kód používáme `PART_NUMBER`; pokud chybí, použijeme `ID`.
- Název skládáme z `PRODUCT` a `DESCRIPTION`.
- Dodavatele posíláme jako `SUPPLIER=HON`.
- Výrobce/značku posíláme v `MANUFACTURER`: ze zdrojové `MAIN_CATEGORY` bezpečně mapujeme `OfficePro` na `Office Pro` a `LÖFFLER` na `LÖFFLER`, ostatní položky zůstávají jako `HON`.
- Dostupnost bereme z `DOSTUPNOST`.
- Obrázky bereme z vnořeného bloku `IMGURL` -> `IMGURL`.
- Doplňkové parametry bereme z `PARAM`.
- Kategorie záměrně neposíláme, aby import nepřepisoval ruční zařazení produktů v Shoptetu.
- Varianty skládáme jen tam, kde jde z `DESCRIPTION` bezpečně určit rozdílné `Provedení`.

### Dřevočal katalog

Adresa vstupu od dodavatele:

```text
https://www.matrace-drevocal.cz/feed-b2b.xml
```

Stažení vstupu od dodavatele:

```bash
curl -L -o source-drevocal.xml https://www.matrace-drevocal.cz/feed-b2b.xml
```

Adresa výstupu parseru:

```text
https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/drevocal.xml
```

Stažení výstupu parseru:

```bash
curl -L -o parser-drevocal.xml https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/drevocal.xml
```

Co děláme:

- Jedna položka `SHOPITEM` ve zdroji odpovídá jedné variantě.
- Varianty skládáme podle `ITEMGROUP_ID`, kód varianty je `ITEM_ID`.
- Kategorii určujeme podle `CATEGORYTEXT`: matrace jdou do `LOŽNICE > MATRACE`, lamelové rošty do `LOŽNICE > ROŠTY`.
- Variantní parametry bereme z `PARAM`; u matrací používáme `Rozměr`, `Výška`, `Potah`, u roštů `Rozměr`.
- Dostupnost bereme z `AVAILABILITY`.
- Popis neposíláme, aby import nepřepisoval ručně spravované popisy.
- Pokud zdroj obsahuje `GIFT`, posíláme ho jako doplňkový parametr `Dárek`.

### Sakypaky katalog

Adresa vstupu od dodavatele:

```text
https://www.sakypaky.cz/export/b2b_partners_cs.xml
```

Stažení vstupu od dodavatele:

```bash
curl -L -o source-sakypaky.xml https://www.sakypaky.cz/export/b2b_partners_cs.xml
```

Adresa výstupu parseru:

```text
https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/sakypaky.xml
```

Stažení výstupu parseru:

```bash
curl -L -o parser-sakypaky.xml https://arpari-xml-feed-parser-railway-production.up.railway.app/feeds/sakypaky.xml
```

Co děláme:

- Posíláme sedací vaky, sedací pytle, taburety, houpačky, stolky, sety, náplně a opravné sady.
- Vynecháváme pelechy a psí produkty, etikety, jmenovky, obalové materiály a nejasné kategorie podle `PRODUCTNAME` a `CATEGORYTEXT`.
- Produkt identifikujeme podle `CODE`.
- Dostupnost odvozujeme z `DELIVERY_DATE`.
- Obrázky bereme z hlavního obrázku a alternativních obrázků (`IMGURL`, `IMGURL_ALTERNATIVE`).
- Varianty skládáme podle `ITEMGROUP_ID`; společná část názvů ve skupině zůstává jako název produktu a rozdílná část jde do variantního parametru `Barva`.
- Variantní parametr `Barva` bereme z `PARAM`, kde `PARAM_NAME` je `Barva` a hodnota je ve `VAL`.
- U konzistentních řad `Dušinka`, kde zdroj posílá hodnoty typu `Dušinka ANTONIE`, ponecháváme `Dušinka` v názvu produktu a do varianty posíláme jen konkrétní motiv, například `ANTONIE`.

## Když se objeví problém

Pokud dodavatel změní strukturu feedu, přestane posílat některé pole, změní kategorie, kódy, variantní parametry nebo obrázky, projeví se to i ve výstupu parseru. Parser data převádí podle pravidel výše, ale chybějící dodavatelská data sám nevytváří.

U každého problému je proto dobré rozlišit, jestli jde o chybu parseru, změnu ve vstupním feedu dodavatele, nebo nastavení importu v Shoptetu.
