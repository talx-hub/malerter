BEGIN TRANSACTION;

CREATE TABLE type(
    id_type INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name_type VARCHAR(32) UNIQUE NOT NULL
);

CREATE TABLE designation(
    id_designation INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name_designation VARCHAR(64) UNIQUE NOT NULL
);

CREATE TABLE metric(
    id_metric INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    delta_metric INT,
    value_metric DOUBLE PRECISION,
    type_metric INT NOT NULL REFERENCES type(id_type),
    name_metric INT NOT NULL REFERENCES designation(id_designation),
    CONSTRAINT unique_type_name UNIQUE (type_metric, name_metric)
);

COMMIT;
