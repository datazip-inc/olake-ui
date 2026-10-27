// The collections TestMongoDB syncs, one per behaviour; see mongodb_test.go for what each checks.
// mongod runs this on its first start, against an empty data directory.
const database = db.getSiblingDB("test_db")

// The same documents twice: Normalization on, so fields become columns, and off, one data column.
const people = [
  { name: "alice", age: 30, address: { city: "paris" }, tags: ["a", "b"] },
  { name: "bob", age: 25, address: { city: "berlin" }, tags: [] },
  { name: "carol", age: 41, address: { city: "rome" }, tags: ["c"] },
]
database.e2e_full_refresh.insertMany(people)
database.e2e_norm_off.insertMany(people)

// Incremental: cursor updated_at, fallback cursor created_at for a document without updated_at.
database.e2e_incremental.insertMany([
  { name: "i1", updated_at: ISODate("2024-01-01T00:00:00Z"), created_at: ISODate("2024-01-01T00:00:00Z") },
  { name: "i2", updated_at: ISODate("2024-01-02T00:00:00Z"), created_at: ISODate("2024-01-02T00:00:00Z") },
  { name: "i3", updated_at: ISODate("2024-01-03T00:00:00Z"), created_at: ISODate("2024-01-03T00:00:00Z") },
])

// Partition regex on category: folders category=a/ and category=b/.
database.e2e_partition.insertMany([
  { name: "p1", category: "a" },
  { name: "p2", category: "a" },
  { name: "p3", category: "b" },
  { name: "p4", category: "b" },
])

// Column selection: secret is deselected when the job is created, note later in Edit Streams.
database.e2e_columns.insertMany([
  { name: "c1", amount: 1, secret: "s1", note: "n1" },
  { name: "c2", amount: 2, secret: "s2", note: "n2" },
])

// Data Filter category = "a" or amount <= 10, on an Incremental stream: f1, f2 and f4 match.
database.e2e_filter_or.insertMany([
  { name: "f1", category: "a", amount: 50, updated_at: ISODate("2024-01-01T00:00:00Z") },
  { name: "f2", category: "b", amount: 5, updated_at: ISODate("2024-01-01T00:00:00Z") },
  { name: "f3", category: "b", amount: 50, updated_at: ISODate("2024-01-01T00:00:00Z") },
  { name: "f4", category: "c", amount: 10, updated_at: ISODate("2024-01-01T00:00:00Z") },
])

// Data Filter amount != 30 and amount >= 20, on an Incremental stream: g1 and g3 match.
database.e2e_filter_and.insertMany([
  { name: "g1", amount: 20, updated_at: ISODate("2024-01-01T00:00:00Z") },
  { name: "g2", amount: 30, updated_at: ISODate("2024-01-01T00:00:00Z") },
  { name: "g3", amount: 40, updated_at: ISODate("2024-01-01T00:00:00Z") },
  { name: "g4", amount: 10, updated_at: ISODate("2024-01-01T00:00:00Z") },
])

// Partition regex on the timestamp event_at, by year and month.
database.e2e_partition_ts.insertMany([
  { name: "t1", event_at: ISODate("2024-01-15T10:00:00Z") },
  { name: "t2", event_at: ISODate("2024-01-20T10:00:00Z") },
  { name: "t3", event_at: ISODate("2025-06-01T10:00:00Z") },
])

// Not in the job at first: Edit Streams adds it.
database.e2e_added.insertMany([
  { name: "x1", age: 1 },
  { name: "x2", age: 2 },
])

// Integer ids, unlike every other collection's ObjectIds: Bulk Configure with it and another
// collection finds no cursor field of one type in both.
database.e2e_int_ids.insertMany([
  { _id: 1, code: 1 },
  { _id: 2, code: 2 },
])
