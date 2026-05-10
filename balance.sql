-- UN POS NO ES MAS QUE UN SISTEMA DE BALANCE(una venta es una transaccion de "stock" -> "cliente" de algun recuro como comida, etc. y de "cliente"->"cuentapos" de otro actuvo, dinero en alguna moneda)
-- A LO SUMO CON UN CONTROL DE STOCK SINCRONIZADO
-- DEBERIA PODER DAR BALANCES..
-- DEBERIA PERMITIR UNA FUNCIONALIDAD DE CUENTA (abrir en tal horario, cerrar en otro, no permitir transacciones desde/hacia dicha cuenta si la cuenta no esta abierta, etc.)
-- DEBERIA ABSTRAERME todo lo que es movimiento de dinero y stock, proveedores, balances, etc. y encapsularlo como un servicio...
CREATE TYPE estado_alta_enum AS ENUM ('ALTA', 'BAJA');
CREATE TYPE estado_cuenta_enum AS ENUM ('ABIERTA', 'CERRADA');
CREATE TYPE estado_transaccion_enum AS ENUM ('PENDIENTE', 'FINALIZADA', 'CANCELADA');

CREATE TABLE Activo(
    id SERIAL PRIMARY KEY,
    creado TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ult_mod TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    estado estado_alta_enum NOT NULL DEFAULT 'ALTA',
    
    nombre VARCHAR(100) UNIQUE NOT NULL,
    unidad VARCHAR(100) NOT NULL,
    UNIQUE(nombre,unidad)
);
-- cuantas unidades del activo origen "vale" el destino, ej orig=pesos(unidad $), destino=oro(unidad kg), recomendado=1000, 1000pesos se intercambian por 1kg de oro
-- Esto es una "recomendacion" para realizar transacciones, pero para nada tiene que generar/forzar movimientos, si se quiere registar una transaccion de 800 pesos por 1kg de oro deberia poderse...
CREATE TABLE Conversion(
    id SERIAL PRIMARY KEY,
    creado TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ult_mod TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    estado estado_alta_enum NOT NULL DEFAULT 'ALTA',
    
    activo_origen INT NOT NULL,
    activo_destino INT NOT NULL,
    recomendado NUMERIC(38,12) NOT NULL DEFAULT 1,
    FOREIGN KEY (activo_origen) REFERENCES Activo(id),
    FOREIGN KEY (activo_destino) REFERENCES Activo(id),
    UNIQUE (activo_origen,activo_destino)
);

CREATE TABLE HistorialConversion(
    reloj TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recomendado_anterior NUMERIC(38,12) NOT NULL,
    conversion_id INT NOT NULL,
    FOREIGN KEY (conversion_id) REFERENCES Conversion(id),
    UNIQUE(conversion_id,reloj)
);

-- Usuario/Empresa id es el id devuelto por el sistema de autenticacion en el token de respuesta, el sistema de balance no deberia generar nunca sus propios usuario_id.
CREATE TABLE Cuenta(
    id SERIAL PRIMARY KEY,
    creado TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ult_mod TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    estado estado_alta_enum NOT NULL DEFAULT 'ALTA',

    permite_deuda BOOLEAN DEFAULT FALSE,
    usuario_id INT NOT NULL, -- en si siempre voy a tener al menos al usuario "root" de la empresa en cuestion...
    empresa_id INT NOT NULL,
    nombre VARCHAR(500) NOT NULL,
    UNIQUE(usuario_id,nombre),
    UNIQUE(empresa_id,nombre)
);

CREATE TABLE MontoCuenta(
    creado TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ult_mod TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cuenta_id INT NOT NULL,
    activo_id INT NOT NULL,
    monto NUMERIC(38,12) NOT NULL DEFAULT 0,
    FOREIGN KEY (cuenta_id) REFERENCES Cuenta(id),
    FOREIGN KEY (activo_id) REFERENCES Activo(id),
    UNIQUE(cuenta_id,activo_id)
);

CREATE TABLE Transaccion(
    id SERIAL PRIMARY KEY,
    creado TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    estado estado_alta_enum NOT NULL DEFAULT 'ALTA',
    
    estado_transaccion estado_transaccion_enum NOT NULL DEFAULT 'PENDIENTE',
    descripcion VARCHAR(255)
);

CREATE TABLE Movimiento (
    id SERIAL PRIMARY KEY,
    
    transaccion_id INT NOT NULL,
    cuenta_id INT NOT NULL,
    activo_id INT NOT NULL,
    monto NUMERIC(38,12) NOT NULL CHECK (monto <> 0),
    
    FOREIGN KEY (transaccion_id) REFERENCES Transaccion(id),
    FOREIGN KEY (cuenta_id) REFERENCES Cuenta(id),
    FOREIGN KEY (activo_id) REFERENCES Activo(id)
);
-- apertura y cierre de cuenta, usuario_id es el responsable del cambio de estado, un usuario diferente puede abrir/cerrar la cuenta de otro (siempre que tenga los permisos)
CREATE TABLE HistorialCuenta(
    id SERIAL PRIMARY KEY,
    
    reloj TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    estado_final estado_cuenta_enum NOT NULL,
    cuenta_id INT NOT NULL,
    usuario_id INT NOT NULL,
    FOREIGN KEY (cuenta_id) REFERENCES Cuenta(id)
);


CREATE OR REPLACE FUNCTION check_transaccion()
RETURNS TRIGGER AS $$
DECLARE
    mov RECORD;
BEGIN
    IF NEW.estado_transaccion <> 'FINALIZADA' THEN
       RETURN NEW;
    END IF;
    IF OLD.estado_transaccion = 'FINALIZADA' THEN
       RAISE EXCEPTION 'Transacción ya finalizada';
    END IF;
    
    IF (SELECT COUNT(*) FROM Movimiento WHERE transaccion_id = NEW.id) < 2 THEN
        RAISE EXCEPTION 'Transacción debe tener al menos 2 movimientos';
    END IF;

    -- BALANCE TRANSACCIONES
    IF EXISTS (
        SELECT 1
        FROM Movimiento
        WHERE transaccion_id = NEW.id
        GROUP BY activo_id
        HAVING ROUND(SUM(monto), 2) <> 0
    ) THEN
         RAISE EXCEPTION 'Transaccion desbalanceada %', NEW.id;
    END IF;

    -- Cuentas abiertas
    IF EXISTS (
        SELECT 1 FROM
        Movimiento m
        JOIN LATERAL (
            SELECT estado_final
            FROM HistorialCuenta h
            WHERE h.cuenta_id = m.cuenta_id
              AND h.reloj <= NEW.creado
            ORDER BY h.reloj DESC
            LIMIT 1
        ) estado ON TRUE
        WHERE transaccion_id = NEW.id
        AND estado.estado_final <> 'ABIERTA'
    ) THEN
        RAISE EXCEPTION 'Hay cuentas cerradas en la transacción %', NEW.id;
    END IF;

    FOR mov IN SELECT * FROM Movimiento WHERE transaccion_id = NEW.id
    LOOP
        INSERT INTO MontoCuenta(cuenta_id,activo_id,monto)
        VALUES (mov.cuenta_id,mov.activo_id,mov.monto)
        ON CONFLICT (cuenta_id, activo_id)
        DO UPDATE SET monto = MontoCuenta.monto + EXCLUDED.monto,ult_mod = CURRENT_TIMESTAMP;
    END LOOP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER transicion_check
BEFORE UPDATE OF estado_transaccion ON Transaccion
FOR EACH ROW
WHEN (NEW.estado_transaccion = 'FINALIZADA')
EXECUTE FUNCTION check_transaccion();


CREATE OR REPLACE FUNCTION estado_cuenta()
RETURNS TRIGGER AS $$
DECLARE
    ultimo_estado estado_cuenta_enum;
BEGIN
    SELECT estado_final
    INTO ultimo_estado
    FROM HistorialCuenta h
    WHERE cuenta_id = NEW.cuenta_id
    ORDER BY h.reloj DESC LIMIT 1;
    IF ultimo_estado IS NOT NULL
    AND ultimo_estado = NEW.estado_final THEN
        RAISE EXCEPTION 'Estado caja invalido %', NEW.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER historial_cuenta_check
BEFORE INSERT ON HistorialCuenta
FOR EACH ROW EXECUTE FUNCTION estado_cuenta();
