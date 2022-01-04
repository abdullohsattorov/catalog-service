package service

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gofrs/uuid"

	pb "github.com/abdullohsattorov/catalog-service/genproto"
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

	book, err := s.storage.Catalog().CreateBook(*req)
	if err != nil {
		s.logger.Error("failed to create book", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to create book")
	}

	return &book, nil
}

func (s *CatalogService) GetBook(ctx context.Context, req *pb.ByIdReq) (*pb.Book, error) {
	book, err := s.storage.Catalog().GetBook(req.GetId())
	if err != nil {
		s.logger.Error("failed to get book", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to get book")
	}

	return &book, nil
}

func (s *CatalogService) ListBook(ctx context.Context, req *pb.ListReq) (*pb.ListRespBook, error) {
	books, count, err := s.storage.Catalog().ListBook(req.Page, req.Limit)
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
	book, err := s.storage.Catalog().UpdateBook(*req)
	if err != nil {
		s.logger.Error("failed to update book", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to update book")
	}

	return &book, nil
}

func (s *CatalogService) DeleteBook(ctx context.Context, req *pb.ByIdReq) (*pb.EmptyResp, error) {
	err := s.storage.Catalog().DeleteBook(req.Id)
	if err != nil {
		s.logger.Error("failed to delete book", l.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete book")
	}

	return &pb.EmptyResp{}, nil
}

func (s *CatalogService) CreateAuthor(ctx context.Context, req *pb.Author) (*pb.Author, error) {
	return req, nil
}

func (s *CatalogService) GetAuthor(ctx context.Context, req *pb.ByIdReq) (*pb.Author, error) {
	return &pb.Author{}, nil
}

func (s *CatalogService) ListAuthor(ctx context.Context, req *pb.ListReq) (*pb.ListRespAuthor, error) {
	return &pb.ListRespAuthor{}, nil
}

func (s *CatalogService) UpdateAuthor(ctx context.Context, req *pb.Author) (*pb.Author, error) {
	return req, nil
}

func (s *CatalogService) DeleteAuthor(ctx context.Context, req *pb.ByIdReq) (*pb.EmptyResp, error) {
	return &pb.EmptyResp{}, nil
}

func (s *CatalogService) CreateCategory(ctx context.Context, req *pb.Category) (*pb.Category, error) {
	return req, nil
}

func (s *CatalogService) GetCategory(ctx context.Context, req *pb.ByIdReq) (*pb.Category, error) {
	return &pb.Category{}, nil
}

func (s *CatalogService) ListCategory(ctx context.Context, req *pb.ListReq) (*pb.ListRespCategory, error) {
	return &pb.ListRespCategory{}, nil
}

func (s *CatalogService) UpdateCategory(ctx context.Context, req *pb.Category) (*pb.Category, error) {
	return req, nil
}

func (s *CatalogService) DeleteCategory(ctx context.Context, req *pb.ByIdReq) (*pb.EmptyResp, error) {
	return &pb.EmptyResp{}, nil
}

func (s *CatalogService) List(ctx context.Context, req *pb.ListReq) (*pb.ListResp, error) {
	return &pb.ListResp{}, nil
}
