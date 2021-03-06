package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gofrs/uuid"

	pb "github.com/abdullohsattorov/catalog-service/genproto/catalog_service"
	l "github.com/abdullohsattorov/catalog-service/pkg/logger"
	"github.com/abdullohsattorov/catalog-service/storage"
)

// CatalogService ...
type CatalogService struct {
	storage storage.IStorage
	logger  l.Logger
}

// NewCatalogService ...
func NewCatalogService(storage storage.IStorage, log l.Logger) *CatalogService {
	return &CatalogService{
		storage: storage,
		logger:  log,
	}
}

func (s *CatalogService) CreateBook(ctx context.Context, req *pb.Book) (*pb.Book, error) {
	id, err := uuid.NewV4()
	if err != nil {
		s.logger.Error("failed while generating uuid", l.Error(err))
		return nil, status.Error(codes.Internal, "failed generate uuid")
	}

	req.BookId = id.String()

	book, err := s.storage.Book().CreateBook(*req)
	if err != nil {
		s.logger.Error("failed to create book", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to create book")
	}

	return &book, nil
}

func (s *CatalogService) GetBook(ctx context.Context, req *pb.ByIdReq) (*pb.Book, error) {
	book, err := s.storage.Book().GetBook(req.GetId())
	if err != nil {
		s.logger.Error("failed to get book", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to get book")
	}

	return &book, nil
}

func (s *CatalogService) ListBook(ctx context.Context, req *pb.ListBookReq) (*pb.ListRespBook, error) {
	books, count, err := s.storage.Book().ListBook(req.Page, req.Limit, req.Filters)
	if err != nil {
		s.logger.Error("failed to list books", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to list books")
	}

	return &pb.ListRespBook{
		Books: books,
		Count: count,
	}, nil
}

func (s *CatalogService) UpdateBook(ctx context.Context, req *pb.Book) (*pb.Book, error) {
	book, err := s.storage.Book().UpdateBook(*req)
	if err != nil {
		s.logger.Error("failed to update book", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to update book")
	}

	return &book, nil
}

func (s *CatalogService) DeleteBook(ctx context.Context, req *pb.ByIdReq) (*pb.EmptyResp, error) {
	err := s.storage.Book().DeleteBook(req.Id)
	if err != nil {
		s.logger.Error("failed to delete book", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete book")
	}

	return &pb.EmptyResp{}, nil
}

func (s *CatalogService) CreateAuthor(ctx context.Context, req *pb.Author) (*pb.Author, error) {
	id, err := uuid.NewV4()
	if err != nil {
		s.logger.Error("failed while generating uuid", l.Error(err))
		return nil, status.Error(codes.Internal, "failed generate uuid")
	}

	req.AuthorId = id.String()

	author, err := s.storage.Author().CreateAuthor(*req)
	if err != nil {
		s.logger.Error("failed to create author", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to create author")
	}
	return &author, nil
}

func (s *CatalogService) GetAuthor(ctx context.Context, req *pb.ByIdReq) (*pb.Author, error) {
	author, err := s.storage.Author().GetAuthor(req.GetId())
	if err != nil {
		s.logger.Error("failed to get author", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to get author")
	}

	return &author, nil
}

func (s *CatalogService) ListAuthor(ctx context.Context, req *pb.ListReq) (*pb.ListRespAuthor, error) {
	authors, count, err := s.storage.Author().ListAuthor(req.Page, req.Limit)
	if err != nil {
		s.logger.Error("failed to list authors", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to list authors")
	}

	return &pb.ListRespAuthor{
		Authors: authors,
		Count:   count,
	}, nil
}

func (s *CatalogService) UpdateAuthor(ctx context.Context, req *pb.Author) (*pb.Author, error) {
	author, err := s.storage.Author().UpdateAuthor(*req)
	if err != nil {
		s.logger.Error("failed to update author", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to update author")
	}
	return &author, nil
}

func (s *CatalogService) DeleteAuthor(ctx context.Context, req *pb.ByIdReq) (*pb.EmptyResp, error) {
	err := s.storage.Author().DeleteAuthor(req.Id)
	if err != nil {
		s.logger.Error("failed to delete author", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete author")
	}
	return &pb.EmptyResp{}, nil
}

func (s *CatalogService) CreateCategory(ctx context.Context, req *pb.Category) (*pb.Category, error) {
	id, err := uuid.NewV4()
	if err != nil {
		s.logger.Error("failed while generating uuid", l.Error(err))
		return nil, status.Error(codes.Internal, "failed generate uuid")
	}

	req.CategoryId = id.String()
	category, err := s.storage.Category().CreateCategory(*req)
	if err != nil {
		s.logger.Error("failed to create category", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to create category")
	}

	return &category, nil
}

func (s *CatalogService) GetCategory(ctx context.Context, req *pb.ByIdReq) (*pb.Category, error) {
	category, err := s.storage.Category().GetCategory(req.GetId())
	if err != nil {
		s.logger.Error("failed to get category", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to get category")
	}
	return &category, nil
}

func (s *CatalogService) ListCategory(ctx context.Context, req *pb.ListReq) (*pb.ListRespCategory, error) {
	categories, count, err := s.storage.Category().ListCategory(req.Page, req.Limit)
	if err != nil {
		s.logger.Error("failed to list categories", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to list categories")
	}

	return &pb.ListRespCategory{
		Categories: categories,
		Count:      count,
	}, nil
}

func (s *CatalogService) UpdateCategory(ctx context.Context, req *pb.Category) (*pb.Category, error) {
	category, err := s.storage.Category().UpdateCategory(*req)
	if err != nil {
		s.logger.Error("failed to update category", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to update category")
	}
	return &category, nil
}

func (s *CatalogService) DeleteCategory(ctx context.Context, req *pb.ByIdReq) (*pb.EmptyResp, error) {
	err := s.storage.Category().DeleteCategory(req.Id)
	if err != nil {
		s.logger.Error("failed to delete category", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete category")
	}

	return &pb.EmptyResp{}, nil
}
