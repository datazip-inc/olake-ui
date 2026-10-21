// Collections the MongoDB workflow syncs, one per behaviour; see mongodb.go for what each is for.
const database = db.getSiblingDB("test_db")

// the same documents twice: normalization on and off
const people = [
  { name: "alice", age: 30, address: { city: "paris" }, tags: ["a", "b"] },
  { name: "bob", age: 25, address: { city: "berlin" }, tags: [] },
  { name: "carol", age: 41, address: { city: "rome" }, tags: ["c"] },
]
database.e2e_full_refresh.insertMany(people)
database.e2e_norm_off.insertMany(people)

// incremental: cursor updated_at, fallback cursor created_at
database.e2e_incremental.insertMany([
  { name: "i1", updated_at: ISODate("2024-01-01T00:00:00Z"), created_at: ISODate("2024-01-01T00:00:00Z") },
  { name: "i2", updated_at: ISODate("2024-01-02T00:00:00Z"), created_at: ISODate("2024-01-02T00:00:00Z") },
  { name: "i3", updated_at: ISODate("2024-01-03T00:00:00Z"), created_at: ISODate("2024-01-03T00:00:00Z") },
])

// partitioned by category into category=a/ and category=b/
database.e2e_partition.insertMany([
  { name: "p1", category: "a" },
  { name: "p2", category: "a" },
  { name: "p3", category: "b" },
  { name: "p4", category: "b" },
])

// not in the first job: the spec adds it through Edit Streams
database.e2e_added.insertMany([
  { name: "x1", age: 1 },
  { name: "x2", age: 2 },
])
