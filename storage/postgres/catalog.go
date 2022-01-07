package postgres

import (
	"database/sql"
	"errors"
	"time"

	sqlbuilder "github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"

	pb "github.com/abdullohsattorov/catalog-service/genproto/catalog_service"
	"github.com/abdullohsattorov/catalog-service/pkg/utils"
)

type catalogRepo struct {
	db *sqlx.DB
}

// NewCatalogRepo ...
func NewCatalogRepo(db *sqlx.DB) *catalogRepo {
	return &catalogRepo{db: db}
}

func (r *catalogRepo) CreateBook(book pb.Book) (pb.Book, error) {
	var id string
	err := r.db.QueryRow(`
        INSERT INTO books(book_id, name, author_id, price, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6) returning book_id`, book.BookId, book.Name, book.AuthorId, book.Price, time.Now().UTC(), time.Now().UTC()).Scan(&id)
	if err != nil {
		return pb.Book{}, err
	}

	_, err = r.db.Exec(`
        INSERT INTO book_categories(book_id, category_id)
        VALUES ($1, $2)`, book.BookId, book.CategoryId)
	if err != nil {
		return pb.Book{}, err
	}

	var NewBook pb.Book

	NewBook, err = r.GetBook(id)

	if err != nil {
		return pb.Book{}, err
	}

	return NewBook, nil
}

func (r *catalogRepo) GetBook(id string) (pb.Book, error) {
	var book pb.Book
	err := r.db.QueryRow(`
        SELECT book_id, name, author_id, price, created_at, updated_at FROM books
        WHERE book_id=$1 and deleted_at is null`, id).Scan(&book.BookId, &book.Name, &book.AuthorId, &book.Price, &book.CreatedAt, &book.UpdatedAt)
	if err != nil {
		return pb.Book{}, err
	}

	err = r.db.QueryRow(`
        select c.category_id, c.name
			from book_categories
		join books b on book_categories.book_id = b.book_id
		join categories c on book_categories.category_id = c.category_id
		where b.book_id = $1`, id).Scan(&book.CategoryId, &book.CategoryName)
	if err != nil {
		return pb.Book{}, err
	}

	return book, nil
}

func (r *catalogRepo) ListBook(page, limit int64) ([]*pb.Book, int64, error) {
	offset := (page - 1) * limit
	rows, err := r.db.Queryx(
		`SELECT book_id, name, author_id, price, created_at, updated_at FROM books WHERE deleted_at is null order by book_id LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	defer rows.Close() // nolint:err check

	var (
		books []*pb.Book
		count int64
	)
	for rows.Next() {
		var book pb.Book
		err = rows.Scan(&book.BookId, &book.Name, &book.AuthorId, &book.Price, &book.CreatedAt, &book.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		err = r.db.QueryRow(`
        select c.category_id, c.name
			from book_categories
		join books b on book_categories.book_id = b.book_id
		join categories c on book_categories.category_id = c.category_id
		where b.book_id = $1`, book.BookId).Scan(&book.CategoryId, &book.CategoryName)
		if err != nil {
			return nil, 0, err
		}
		books = append(books, &book)
	}

	err = r.db.QueryRow(`SELECT count(*) FROM books where deleted_at is null`).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return books, count, nil
}

func (r *catalogRepo) UpdateBook(book pb.Book) (pb.Book, error) {
	result, err := r.db.Exec(`UPDATE books SET name=$1, author_id=$2, price=$3, updated_at = $4 WHERE book_id=$5 and deleted_at is null`,
		book.Name, book.AuthorId, book.Price, time.Now().UTC(), book.BookId,
	)
	if err != nil {
		return pb.Book{}, err
	}

	if i, _ := result.RowsAffected(); i == 0 {
		return pb.Book{}, sql.ErrNoRows
	}

	resultNext, err := r.db.Exec(`UPDATE book_categories SET category_id = $1 WHERE book_id=$2`,
		book.CategoryId, book.BookId,
	)
	if err != nil {
		return pb.Book{}, err
	}

	if i, _ := resultNext.RowsAffected(); i == 0 {
		return pb.Book{}, sql.ErrNoRows
	}

	var NewBook pb.Book

	NewBook, err = r.GetBook(book.BookId)

	if err != nil {
		return pb.Book{}, err
	}

	return NewBook, nil
}

func (r *catalogRepo) DeleteBook(id string) error {
	result, err := r.db.Exec(`UPDATE books SET deleted_at = $1 WHERE book_id=$2`, time.Now().UTC(), id)
	if err != nil {
		return err
	}

	if i, _ := result.RowsAffected(); i == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *catalogRepo) CreateCategory(category pb.Category) (pb.Category, error) {
	var id string
	if category.ParentUuid != "" {
		err := r.db.QueryRow(`
		INSERT INTO categories(category_id, name, parent_uuid, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5) returning category_id`, category.CategoryId, category.Name, category.ParentUuid, time.Now().UTC(), time.Now().UTC()).Scan(&id)
		if err != nil {
			return pb.Category{}, err
		}
	} else {
		err := r.db.QueryRow(`
		INSERT INTO categories(category_id, name, created_at, updated_at) 
		VALUES ($1, $2, $3, $4) returning category_id`, category.CategoryId, category.Name, time.Now().UTC(), time.Now().UTC()).Scan(&id)
		if err != nil {
			return pb.Category{}, err
		}
	}

	newCategory, err := r.GetCategory(id)
	if err != nil {
		return pb.Category{}, err
	}

	return newCategory, nil
}

func (r *catalogRepo) GetCategory(id string) (pb.Category, error) {
	var category pb.Category
	var (
		parentUUID     sql.NullString
		parentCategory sql.NullString
	)

	err := r.db.QueryRow(`
		SELECT cat.category_id, cat.name AS category_name, cat.parent_uuid, cat2.name AS parent_category, cat.created_at, cat.updated_at
		FROM categories AS cat 
		LEFT JOIN categories AS cat2 ON cat.parent_uuid = cat2.category_id
		WHERE cat.category_id=$1 AND cat.deleted_at is null;
		`, id).Scan(&category.CategoryId, &category.Name, &parentUUID, &parentCategory, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return pb.Category{}, err
	}

	if !parentUUID.Valid {
		parentUUID.String = ""
	}
	category.ParentUuid = parentUUID.String

	if !parentCategory.Valid {
		parentCategory.String = ""
	}
	category.ParentCategory = parentCategory.String

	return category, nil
}

func (r *catalogRepo) ListCategory(page, limit int64) ([]*pb.Category, int64, error) {
	offset := (page - 1) * limit
	rows, err := r.db.Queryx(`
		SELECT cat.category_id, cat.name AS category_name, cat.parent_uuid, cat2.name AS parent_category, cat.created_at, cat.updated_at
		FROM categories AS cat 
		LEFT JOIN categories AS cat2 ON cat.parent_uuid = cat2.category_id
		WHERE cat.deleted_at is null 
		ORDER BY category_id LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var (
		categories []*pb.Category
		count      int64
	)
	var (
		parentUUID     sql.NullString
		parentCategory sql.NullString
	)
	for rows.Next() {
		var category pb.Category

		err = rows.Scan(&category.CategoryId, &category.Name, &parentUUID, &parentCategory, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}

		if !parentUUID.Valid {
			parentUUID.String = ""
		}
		category.ParentUuid = parentUUID.String

		if !parentCategory.Valid {
			parentCategory.String = ""
		}
		category.ParentCategory = parentCategory.String

		categories = append(categories, &category)
	}

	err = r.db.QueryRow(`SELECT count(*) FROM categories where deleted_at is null`).Scan(&count)

	if err != nil {
		return nil, 0, err
	}

	return categories, count, nil
}

func (r *catalogRepo) UpdateCategory(category pb.Category) (pb.Category, error) {
	if category.ParentUuid != "" {
		result, err := r.db.Exec(`UPDATE categories SET name=$1, parent_uuid=$2, updated_at = $3 WHERE category_id=$4 and deleted_at is null`,
			category.Name, category.ParentUuid, time.Now().UTC(), category.CategoryId,
		)
		if err != nil {
			return pb.Category{}, err
		}
		if i, _ := result.RowsAffected(); i == 0 {
			return pb.Category{}, sql.ErrNoRows
		}
	} else {
		result, err := r.db.Exec(`UPDATE categories SET name=$1, parent_uuid=null, updated_at = $2 WHERE category_id=$3 and deleted_at is null`,
			category.Name, time.Now().UTC(), category.CategoryId,
		)
		if err != nil {
			return pb.Category{}, err
		}
		if i, _ := result.RowsAffected(); i == 0 {
			return pb.Category{}, sql.ErrNoRows
		}
	}

	var newCategory pb.Category

	newCategory, err := r.GetCategory(category.CategoryId)
	if err != nil {
		return pb.Category{}, err
	}

	return newCategory, nil
}

func (r *catalogRepo) DeleteCategory(id string) error {
	var count int
	err := r.db.QueryRow(`select count(*) from categories where parent_uuid=$1 `, id).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		err1 := errors.New("this category has subcategories")
		return err1
	}

	result, err := r.db.Exec(`UPDATE categories SET deleted_at = $1 WHERE category_id=$2`, time.Now().UTC(), id)
	if err != nil {
		return err
	}

	if i, _ := result.RowsAffected(); i == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *catalogRepo) CreateAuthor(author pb.Author) (pb.Author, error) {
	var id string

	err := r.db.QueryRow(`
				   INSERT INTO authors (author_id, name, created_at, updated_at) 
				   VALUES ($1, $2, $3, $4) RETURNING author_id `,
		author.AuthorId, author.Name, time.Now().UTC(), time.Now().UTC()).
		Scan(&id)
	if err != nil {
		return pb.Author{}, err
	}

	NewAuthor, err := r.GetAuthor(id)
	if err != nil {
		return pb.Author{}, err
	}

	return NewAuthor, nil
}

func (r *catalogRepo) GetAuthor(id string) (pb.Author, error) {
	var NewAuthor pb.Author

	err := r.db.QueryRow(`
						SELECT author_id, name, created_at, updated_at FROM authors 
						WHERE author_id = $1 AND deleted_at IS NULL`, id).
		Scan(&NewAuthor.AuthorId, &NewAuthor.Name, &NewAuthor.CreatedAt, &NewAuthor.UpdatedAt)
	if err != nil {
		return pb.Author{}, err
	}

	return NewAuthor, nil
}

func (r *catalogRepo) ListAuthor(page, limit int64) ([]*pb.Author, int64, error) {
	offset := (page - 1) * limit

	rows, err := r.db.Queryx(`
				SELECT author_id, name, created_at, updated_at FROM authors 
				WHERE deleted_at is NULL ORDER BY author_id LIMIT $1 OFFSET $2
				`, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var (
		authors []*pb.Author
		count   int64
	)

	for rows.Next() {
		var author pb.Author
		err = rows.Scan(&author.AuthorId, &author.Name, &author.CreatedAt, &author.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		authors = append(authors, &author)
	}

	err = r.db.QueryRow("SELECT count(*) FROM authors WHERE deleted_at IS NULL").Scan(&count)
	if err != nil {
		return nil, 0, err
	}
	return authors, count, nil
}

func (r *catalogRepo) UpdateAuthor(update pb.Author) (pb.Author, error) {
	result, err := r.db.Exec("UPDATE authors SET name = $2, updated_at = $3 WHERE author_id = $1",
		update.AuthorId, update.Name, time.Now().UTC())
	if err != nil {
		return pb.Author{}, err
	}

	if i, _ := result.RowsAffected(); i == 0 {
		return pb.Author{}, sql.ErrNoRows
	}

	var NewAuthor pb.Author

	NewAuthor, err = r.GetAuthor(update.AuthorId)

	if err != nil {
		return pb.Author{}, err
	}
	return NewAuthor, nil
}

func (r *catalogRepo) DeleteAuthor(id string) error {
	result, err := r.db.Exec("UPDATE authors SET deleted_at = $2 WHERE author_id = $1", id, time.Now().UTC())
	if err != nil {
		return err
	}

	if i, _ := result.RowsAffected(); i == 0 {
		return sql.ErrNoRows
	}

	return nil
}

//func (r *catalogRepo) List(page, limit int64, filters map[string]string) ([]*pb.Catalog, int64, error) {
//	offset := (page - 1) * limit
//	var args string
//	if value, ok := filters["authors"]; ok {
//		args = value
//	}
//	rows, err := r.db.Queryx(`
//		select
//			b.book_id,
//			b.name,
//			b.price,
//			b.created_at,
//			b.updated_at,
//			a.author_id,
//			a.name,
//			a.created_at,
//			a.updated_at
//		from
//			book_categories
//		join books b on book_categories.book_id = b.book_id
//		join categories c on book_categories.category_id = c.category_id
//		join authors a on b.author_id = a.author_id
//		where b.deleted_at is null and b.author_id in ($1)
//		LIMIT $2 OFFSET $3`,
//		args, limit, offset)
//	if err != nil {
//		return nil, 0, err
//	}
//
//	var (
//		catalogs []*pb.Catalog
//		count    int64
//	)
//
//	_, count, _ = r.ListBook(1, 1)
//
//	for rows.Next() {
//		var catalog pb.Catalog
//		var book pb.Book
//		var author pb.Author
//		err = rows.Scan(
//			&book.BookId, &book.Name, &book.Price, &book.CreatedAt, &book.UpdatedAt,
//			&author.AuthorId, &author.Name, &author.CreatedAt, &author.UpdatedAt,
//		)
//		if err != nil {
//			return nil, 0, err
//		}
//
//		rowsCategory, err := r.db.Queryx(`
//			select
//				c.category_id,
//				c.name,
//				c.parent_uuid,
//				c.created_at,
//				c.updated_at
//			from
//				book_categories
//					join books b on book_categories.book_id = b.book_id
//					join categories c on book_categories.category_id = c.category_id
//			where b.deleted_at is null and b.book_id = $1`, book.BookId)
//		if err != nil {
//			return nil, 0, err
//		}
//
//		for rowsCategory.Next() {
//			var category pb.Category
//			err = rowsCategory.Scan(
//				&category.CategoryId, &category.Name, &category.ParentUuid, &category.CreatedAt, &category.UpdatedAt,
//			)
//			catalog.Category = append(catalog.Category, &category)
//		}
//
//		if err != nil {
//			return nil, 0, err
//		}
//
//		catalog.Book = &book
//		catalog.Author = &author
//		catalogs = append(catalogs, &catalog)
//	}
//
//	return catalogs, count, nil
//}

func (r *catalogRepo) List(page, limit int64, filters map[string]string) ([]*pb.Catalog, int64, error) {
	offset := (page - 1) * limit
	sb := sqlbuilder.NewSelectBuilder()

	sb.Select("b.book_id", "b.name", "b.price", "b.created_at", "b.updated_at", "a.author_id", "a.name", "a.created_at", "a.updated_at")
	sb.From("book_categories")
	sb.Join("books b", "book_categories.book_id=b.book_id")
	sb.Join("categories c", "book_categories.category_id=c.category_id")
	sb.Join("authors a", "b.author_id=a.author_id")
	sb.GroupBy("b.book_id, a.author_id")
	if value, ok := filters["authors"]; ok {
		args := utils.StringSliceToInterfaceSlice(utils.ParseFilter(value))
		sb.Where(sb.In("a.author_id", args...))
	}

	if value, ok := filters["categories"]; ok {
		args := utils.StringSliceToInterfaceSlice(utils.ParseFilter(value))
		sb.Where(sb.In("c.category_id", args...))
	}

	sb.Limit(int(limit))
	sb.Offset(int(offset))
	query, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	rows, err := r.db.Queryx(query, args...)
	if err != nil {
		return nil, 0, err
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var (
		catalogs []*pb.Catalog
		count    int64
	)

	for rows.Next() {
		var catalog pb.Catalog
		var book pb.Book
		var author pb.Author
		err = rows.Scan(
			&book.BookId, &book.Name, &book.Price, &book.CreatedAt, &book.UpdatedAt,
			&author.AuthorId, &author.Name, &author.CreatedAt, &author.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		cs := sqlbuilder.NewSelectBuilder()

		cs.Select("c.category_id", "c.name", "c.parent_uuid", "c.created_at", "c.updated_at")
		cs.Join("books b", "book_categories.book_id=b.book_id")
		cs.Join("categories c", "book_categories.category_id=c.category_id")
		cs.Where(cs.Equal("b.book_id", book.BookId))
		cs.Where("b.deleted_at is null")
		cs.From("book_categories")

		query, args := cs.BuildWithFlavor(sqlbuilder.PostgreSQL)

		rowsCategory, err := r.db.Queryx(query, args...)
		if err != nil {
			return nil, 0, err
		}

		for rowsCategory.Next() {
			var category pb.Category
			err = rowsCategory.Scan(
				&category.CategoryId, &category.Name, &category.ParentUuid, &category.CreatedAt, &category.UpdatedAt,
			)
			if err != nil {
				return nil, 0, err
			}
			catalog.Category = append(catalog.Category, &category)
		}

		catalog.Book = &book
		catalog.Author = &author
		catalogs = append(catalogs, &catalog)
	}

	return catalogs, count, nil
}
