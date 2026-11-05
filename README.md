# DDD Tavern

A Domain-Driven Design learning project written in Go. The application models a
small tavern where customers can place orders and be billed for the products
they buy. It was built to practise DDD building blocks, repository boundaries,
dependency injection through service configuration, and the integration of the
same domain repositories with in-memory, MongoDB, and PostgreSQL persistence.

> This is a learning project, not a production application. It intentionally
> focuses on the domain model, repository implementations, and service
> workflows. No live database connection or environment configuration is
> required.

## Domain Model

### Entities

- **Person** – a person with an ID, name, and age.
- **Item** – a product item with an ID, name, and description.

### Value Object

- **Transaction** – an immutable record of a payment amount, a source customer,
  a destination account, and the time the transaction was created.

### Aggregates

- **Customer** – the customer aggregate root. It wraps a `Person` and is
  created through a factory that rejects empty names. It exposes its ID and name
  and supports restoring an ID when loading from a repository.
- **Product** – the product aggregate root. It wraps an `Item` and stores the
  product price and quantity. It is created through a factory that requires a
  name and description, and it exposes setters and getters used when mapping
  between the domain and a database.

### Repository Interfaces

The domain defines persistence contracts without depending on a database:

- **CustomerRepository** – `Get`, `Add`, and `Update` customers, with domain
  errors for missing customers, failed inserts, and failed updates.
- **ProductRepository** – `GetAll`, `GetByID`, `Add`, `Update`, and `Delete`
  products, with domain errors for missing products and duplicate products.

## Persistence Implementations

### In-Memory

The original repositories store aggregates in `map` values keyed by UUID and use
a mutex for safe concurrent access. They provide the baseline implementation
used by the services during normal development.

### MongoDB

MongoDB repositories are provided for both customers and products. Each
repository receives a MongoDB collection and maps the aggregate to an explicit
BSON document:

- Customers are stored in the `customers` collection with `_id` and `name`.
- Products are stored in the `products` collection with `_id`, `name`,
  `description`, `price`, and `quantity`.

The MongoDB implementations support the full repository interface: customer
get/add/update and product list/get/add/update/delete. MongoDB's "no documents"
result is translated into the domain's not-found errors, and duplicate-key
errors are translated into the domain's duplicate/insert errors. The repository
does not create a connection itself; the caller supplies the collection.

### PostgreSQL

PostgreSQL repositories are provided for both customers and products using
standard SQL queries and PostgreSQL placeholders. The schema contains:

- `customers` – UUID primary key and customer name.
- `products` – UUID primary key, name, description, price, and quantity.

The PostgreSQL implementations provide the same repository behavior as the
in-memory and MongoDB versions. They translate `sql.ErrNoRows` into domain
not-found errors, PostgreSQL unique-constraint violations (`23505`) into
duplicate/insert errors, and zero affected rows into update/delete errors. The
repository receives an existing `*sql.DB` and does not manage connection
configuration.

The SQL schema is available in `migrations/001_init.sql`.

## Services and Workflows

The service layer is wired with functional configuration options. Each service
accepts repository implementations, so the same workflow can run against
in-memory, MongoDB, or PostgreSQL repositories.

### Order Service

The order workflow:

1. Load the customer by ID.
2. Load every requested product by ID.
3. Accumulate the product prices into a total.
4. Record the customer and the number of ordered products in the application
   log.

The service exposes a customer repository option and a memory-backed product
repository option.

### Tavern Service

The tavern workflow performs the same customer and product validation as the
order service. It calculates the total price of the ordered products and logs
the customer, the product count, and the total value.

### Billing Service

The billing workflow:

1. Load the customer by ID.
2. Load every requested product by ID.
3. Calculate the total amount from product prices.
4. Create a `Transaction` value object from the customer to the tavern.
5. Log the billed amount and return the transaction to the caller.

## Functionality Summary

- Create customers with validated names.
- Retrieve and update customers through the repository contract.
- Create products with validated name and description.
- List, retrieve, create, update, and delete products through the repository
  contract.
- Place customer orders and validate that both the customer and products exist.
- Calculate order totals from product prices.
- Bill customers for ordered products and return a transaction value object.
- Persist and restore customers and products with in-memory, MongoDB, or
  PostgreSQL repositories.
- Keep the domain independent of any specific database through repository
  interfaces.
- Exercise repository behavior without a live database by supplying MongoDB
  mock deployments and SQL mock connections in the repository tests.

The project ends here as a completed Domain-Driven Design learning exercise.
