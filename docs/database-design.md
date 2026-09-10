# Проєкт бази даних Fire_depart

СУБД: **PostgreSQL 16**. Схема: `db/migrations/001_initial_schema.sql`.

## Сутності та призначення

| Таблиця | Призначення |
|---|---|
| `fire_stations` | Пожежні части (назва, адреса, телефон) |
| `ranks` | Довідник звань |
| `employees` | Особовий склад; належить части, має звання |
| `shifts` | Зміни (частина, час початку/закінчення) |
| `shift_assignments` | Призначення працівників на зміни (M:N) |
| `vehicles` | Техніка (позивний, тип, статус) |
| `incident_types` | Довідник типів викликів |
| `incidents` | Виклики: адреса, опис, час отримання/закриття, статус |
| `incident_vehicles` | Яка техніка виїжджала на виклик (M:N) |
| `incident_responders` | Хто з працівників брав участь (M:N) |
| `incident_reports` | Підсумок виклику: опис дій, збитки (грн), постраждалі; 1:1 до виклику |
| `schema_layouts` | Збережені позиції вузлів інтерактивної ERD-мапи |

## Зв'язки (зовнішні ключі)

- `employees.station_id → fire_stations.id` (RESTRICT), `employees.rank_id → ranks.id` (SET NULL)
- `shifts.station_id → fire_stations.id` (CASCADE)
- `shift_assignments.shift_id → shifts.id`, `.employee_id → employees.id` (CASCADE)
- `vehicles.station_id → fire_stations.id` (RESTRICT)
- `incidents.station_id → fire_stations.id`, `.type_id → incident_types.id` (RESTRICT)
- `incident_vehicles.incident_id → incidents.id` (CASCADE), `.vehicle_id → vehicles.id` (RESTRICT)
- `incident_responders.incident_id → incidents.id` (CASCADE), `.employee_id → employees.id` (RESTRICT)
- `incident_reports.incident_id → incidents.id` (CASCADE, UNIQUE → 1:1)

Проміжні таблиці `shift_assignments`, `incident_vehicles`, `incident_responders` реалізують зв'язки «багато-до-багатьох» із `UNIQUE`-обмеженнями на пари, що виключає дублікати.

## Обмеження цілісності

- `CHECK (ends_at > starts_at)` — зміна не може закінчитися раніше початку
- `CHECK (closed_at IS NULL OR closed_at >= received_at)` — закриття виклику не раніше отримання
- `CHECK (status IN (...))` — контроль допустимих статусів викликів і техніки
- `CHECK (damage_uah >= 0)`, `CHECK (casualties >= 0)` — невід'ємні збитки/постраждалі
- `ON DELETE`-політики: CASCADE для дочірніх записів виклику, RESTRICT для критичних посилань (частина, техніка, працівник)

## Індекси

- `employees(station_id)`, `incidents(station_id, status)`, `incidents(received_at)`, `vehicles(station_id, status)` — під типові запити дашборда та звітів.
