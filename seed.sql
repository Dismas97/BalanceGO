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

INSERT INTO TipoTransaccion (nombre) VALUES
('DEPOSITO'),
('RETIRO'),
('VENTA'),
('GASTO');


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


INSERT INTO Cuenta (nombre,permite_deuda,usuario_id,empresa_id) VALUES
('INTERNO',FALSE,2,2),
('INTERNO:CAJA',FALSE,2,2),
('INTERNO:INVENTARIO',FALSE,2,2),


('EXTERNO',TRUE,2,2),
('EXTERNO:INGRESO:CAJA:ANONIMO',TRUE,2,2),
('EXTERNO:INGRESO:CAJA:EMPRESA',TRUE,2,2),

('EXTERNO:INGRESO:INVENTARIO:ANONIMO',TRUE,2,2),
('EXTERNO:INGRESO:INVENTARIO:EMPRESA',TRUE,2,2),

('EXTERNO:EGRESO:CAJA:ANONIMO',TRUE,2,2),
('EXTERNO:EGRESO:CAJA:EMPRESA',TRUE,2,2),
('EXTERNO:EGRESO:INVENTARIO:ANONIMO',TRUE,2,2),
('EXTERNO:EGRESO:INVENTARIO:EMPRESA',TRUE,2,2),

('EXTERNO:EGRESO:GASTO:CAJA:EMPRESA',TRUE,2,2),
('EXTERNO:EGRESO:GASTO:INVENTARIO:EMPRESA',TRUE,2,2);

INSERT INTO Activo (nombre,unidad_id,empresa_id) values ('Peso Argentino', 4, 2);
