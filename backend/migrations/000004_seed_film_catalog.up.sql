INSERT INTO film_color_types (code, name)
VALUES
    ('color_negative', 'Color Negative'),
    ('black_and_white', 'Black & White'),
    ('slide', 'Slide');


INSERT INTO film_formats (code, name)
VALUES
    ('35mm', '35 mm'),
    ('120', '120'),
    ('220', '220'),
    ('4x5', '4×5'),
    ('8x10', '8×10');


INSERT INTO film_stocks (
    manufacturer,
    name,
    iso,
    color_type_id
)
VALUES
    (
        'Kodak',
        'Gold 200',
        200,
        (SELECT id FROM film_color_types WHERE code = 'color_negative')
    ),
    (
        'Kodak',
        'ColorPlus 200',
        200,
        (SELECT id FROM film_color_types WHERE code = 'color_negative')
    ),
    (
        'Kodak',
        'UltraMax 400',
        400,
        (SELECT id FROM film_color_types WHERE code = 'color_negative')
    ),
    (
        'Kodak',
        'Portra 160',
        160,
        (SELECT id FROM film_color_types WHERE code = 'color_negative')
    ),
    (
        'Kodak',
        'Portra 400',
        400,
        (SELECT id FROM film_color_types WHERE code = 'color_negative')
    ),
    (
        'Kodak',
        'Portra 800',
        800,
        (SELECT id FROM film_color_types WHERE code = 'color_negative')
    ),
    (
        'Kodak',
        'Ektar 100',
        100,
        (SELECT id FROM film_color_types WHERE code = 'color_negative')
    ),
    (
        'Kodak',
        'Ektachrome E100',
        100,
        (SELECT id FROM film_color_types WHERE code = 'slide')
    ),

    (
        'Ilford',
        'HP5 Plus',
        400,
        (SELECT id FROM film_color_types WHERE code = 'black_and_white')
    ),
    (
        'Ilford',
        'FP4 Plus',
        125,
        (SELECT id FROM film_color_types WHERE code = 'black_and_white')
    ),
    (
        'Ilford',
        'Delta 100',
        100,
        (SELECT id FROM film_color_types WHERE code = 'black_and_white')
    ),
    (
        'Ilford',
        'Delta 400',
        400,
        (SELECT id FROM film_color_types WHERE code = 'black_and_white')
    ),
    (
        'Ilford',
        'Delta 3200',
        3200,
        (SELECT id FROM film_color_types WHERE code = 'black_and_white')
    ),

    (
        'Kentmere',
        'Pan 100',
        100,
        (SELECT id FROM film_color_types WHERE code = 'black_and_white')
    ),
    (
        'Kentmere',
        'Pan 400',
        400,
        (SELECT id FROM film_color_types WHERE code = 'black_and_white')
    ),

    (
        'Fujifilm',
        '200',
        200,
        (SELECT id FROM film_color_types WHERE code = 'color_negative')
    ),
    (
        'Fujifilm',
        '400',
        400,
        (SELECT id FROM film_color_types WHERE code = 'color_negative')
    ),
    (
        'Fujifilm',
        'Velvia 50',
        50,
        (SELECT id FROM film_color_types WHERE code = 'slide')
    ),
    (
        'Fujifilm',
        'Provia 100F',
        100,
        (SELECT id FROM film_color_types WHERE code = 'slide')
    ),

    (
        'Kodak',
        'Vision3 50D',
        50,
        (SELECT id FROM film_color_types WHERE code = 'color_negative')
    ),
    (
        'Kodak',
        'Vision3 250D',
        250,
        (SELECT id FROM film_color_types WHERE code = 'color_negative')
    ),
    (
        'Kodak',
        'Vision3 500T',
        500,
        (SELECT id FROM film_color_types WHERE code = 'color_negative')
    ),
    (
        'Foma',
        'Fomapan 100 Classic',
        100,
        (SELECT id FROM film_color_types WHERE code = 'black_and_white')
    ),
    (
        'Foma',
        'Fomapan 200 Creative',
        200,
        (SELECT id FROM film_color_types WHERE code = 'black_and_white')
    ),
    (
        'Foma',
        'Fomapan 400 Action',
        400,
        (SELECT id FROM film_color_types WHERE code = 'black_and_white')
    );



-- Kodak consumer color films: 35 mm
INSERT INTO film_stock_formats (film_stock_id, format_id)
SELECT fs.id, ff.id
FROM film_stocks fs
CROSS JOIN film_formats ff
WHERE fs.manufacturer = 'Kodak'
  AND fs.name IN (
      'Gold 200',
      'ColorPlus 200',
      'UltraMax 400'
  )
  AND ff.code = '35mm';


-- Kodak Portra / Ektar: 35 mm + 120
INSERT INTO film_stock_formats (film_stock_id, format_id)
SELECT fs.id, ff.id
FROM film_stocks fs
CROSS JOIN film_formats ff
WHERE fs.manufacturer = 'Kodak'
  AND fs.name IN (
      'Portra 160',
      'Portra 400',
      'Portra 800',
      'Ektar 100'
  )
  AND ff.code IN ('35mm', '120');


-- Ektachrome
INSERT INTO film_stock_formats (film_stock_id, format_id)
SELECT fs.id, ff.id
FROM film_stocks fs
CROSS JOIN film_formats ff
WHERE fs.manufacturer = 'Kodak'
  AND fs.name = 'Ektachrome E100'
  AND ff.code IN ('35mm', '120');


-- Ilford
INSERT INTO film_stock_formats (film_stock_id, format_id)
SELECT fs.id, ff.id
FROM film_stocks fs
CROSS JOIN film_formats ff
WHERE fs.manufacturer = 'Ilford'
  AND fs.name IN (
      'HP5 Plus',
      'FP4 Plus',
      'Delta 100',
      'Delta 400'
  )
  AND ff.code IN ('35mm', '120');


INSERT INTO film_stock_formats (film_stock_id, format_id)
SELECT fs.id, ff.id
FROM film_stocks fs
CROSS JOIN film_formats ff
WHERE fs.manufacturer = 'Ilford'
  AND fs.name = 'Delta 3200'
  AND ff.code IN ('35mm', '120');


-- Kentmere
INSERT INTO film_stock_formats (film_stock_id, format_id)
SELECT fs.id, ff.id
FROM film_stocks fs
CROSS JOIN film_formats ff
WHERE fs.manufacturer = 'Kentmere'
  AND fs.name IN ('Pan 100', 'Pan 400')
  AND ff.code IN ('35mm', '120');


-- Fujifilm
INSERT INTO film_stock_formats (film_stock_id, format_id)
SELECT fs.id, ff.id
FROM film_stocks fs
CROSS JOIN film_formats ff
WHERE fs.manufacturer = 'Fujifilm'
  AND fs.name IN ('200', '400')
  AND ff.code = '35mm';


INSERT INTO film_stock_formats (film_stock_id, format_id)
SELECT fs.id, ff.id
FROM film_stocks fs
CROSS JOIN film_formats ff
WHERE fs.manufacturer = 'Fujifilm'
  AND fs.name IN ('Velvia 50', 'Provia 100F')
  AND ff.code IN ('35mm', '120');


INSERT INTO film_stock_formats (film_stock_id, format_id)
SELECT fs.id, ff.id
FROM film_stocks fs
CROSS JOIN film_formats ff
WHERE fs.manufacturer = 'Kodak'
  AND fs.name IN (
      'Vision3 50D',
      'Vision3 250D',
      'Vision3 500T'
  )
  AND ff.code IN ('35mm', '120');

INSERT INTO film_stock_formats (film_stock_id, format_id)
SELECT fs.id, ff.id
FROM film_stocks fs
CROSS JOIN film_formats ff
WHERE fs.manufacturer = 'Foma'
  AND fs.name IN (
      'Fomapan 100 Classic',
      'Fomapan 200 Creative',
      'Fomapan 400 Action'
  )
  AND ff.code IN ('35mm', '120');