module backend

go 1.26.3

require backend/shared v0.0.0

require github.com/go-chi/chi/v5 v5.3.2

replace backend/shared => ./shared
