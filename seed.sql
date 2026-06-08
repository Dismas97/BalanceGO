INSERT INTO TipoUnidad (nombre) VALUES
('MASA'),
('VOLUMEN'),
('LONGITUD'),
('AREA'),
('ENERGIA'),
('POTENCIA'),
('FLUJO'),
('CONTABLE DISCRETA'),
('MONETARIA');




INSERT INTO Unidad (nombre, simbolo, tipo_unidad_id)
SELECT 'gramo', 'g', id
FROM TipoUnidad WHERE nombre = 'MASA';

INSERT INTO Unidad (nombre, simbolo, tipo_unidad_id)
SELECT 'kilogramo', 'kg', id
FROM TipoUnidad WHERE nombre = 'MASA';

INSERT INTO Unidad (nombre, simbolo, tipo_unidad_id)
SELECT 'tonelada', 't', id
FROM TipoUnidad WHERE nombre = 'MASA';



INSERT INTO Unidad (nombre, simbolo, tipo_unidad_id)
SELECT 'peso argentino', 'ARS', id
FROM TipoUnidad WHERE nombre = 'MONETARIA';

INSERT INTO Unidad (nombre, simbolo, tipo_unidad_id)
SELECT 'centavo dolar', 'cent_usd', id
FROM TipoUnidad WHERE nombre = 'MONETARIA';
