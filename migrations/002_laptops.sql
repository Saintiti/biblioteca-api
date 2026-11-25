-- Proyecto final: inventario de equipos físicos y sus préstamos.
-- Se ejecuta automáticamente al inicializar una base MySQL nueva.
-- Para un volumen ya existente, ejecutar manualmente esta migración UNA vez.
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS laptops (
  id INT AUTO_INCREMENT PRIMARY KEY,
  serial_number VARCHAR(80) NOT NULL UNIQUE,
  brand VARCHAR(100) NOT NULL,
  model VARCHAR(120) NOT NULL,
  status ENUM('available', 'borrowed') NOT NULL DEFAULT 'available',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS laptop_loans (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  laptop_id INT NOT NULL,
  loaned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  due_at TIMESTAMP NOT NULL,
  returned_at TIMESTAMP NULL DEFAULT NULL,
  status ENUM('active', 'returned') NOT NULL DEFAULT 'active',
  -- Único solo cuando el préstamo está activo: los históricos devueltos admiten NULL.
  active_laptop_id INT GENERATED ALWAYS AS (
    CASE WHEN status = 'active' THEN laptop_id ELSE NULL END
  ) STORED,
  CONSTRAINT fk_laptop_loans_user FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_laptop_loans_laptop FOREIGN KEY (laptop_id) REFERENCES laptops(id),
  UNIQUE KEY uq_laptop_one_active_loan (active_laptop_id),
  INDEX idx_laptop_loans_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
