package service_test

import (
	"errors"
	"github.com/autodidaddict/go-shopping/catalog/internal/service"
	"github.com/autodidaddict/go-shopping/catalog/proto"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
	"net/http"
	"testing"
)

func TestProductRetrieval(t *testing.T) {
	Convey("Given a catalog service", t, func() {
		repo := newFakeRepo()
		svc := service.NewCatalogService(repo)
		ctx := context.Background()

		Convey("Querying for a single product should invoke repository", func() {
			repo.shouldFail = false
			var resp catalog.DetailResponse
			err := svc.GetProductDetails(ctx, &catalog.DetailRequest{SKU: "8675309"}, &resp)
			So(err, ShouldBeNil)
			So(resp.Product.SKU, ShouldEqual, "8675309")
		})

		Convey("Querying for a non-existent product should produce a hinted failure", func() {
			repo.shouldFail = false
			var resp catalog.DetailResponse
			err := svc.GetProductDetails(ctx, &catalog.DetailRequest{SKU: "DONTEXIST"}, &resp)
			So(err, ShouldBeNil)
			So(resp.Error, ShouldNotBeNil)
			So(resp.Error.HttpHint, ShouldEqual, http.StatusNotFound)
		})

		Convey("Querying for a product should fail when repository fails", func() {
			repo.shouldFail = true
			var resp catalog.DetailResponse
			err := svc.GetProductDetails(ctx, &catalog.DetailRequest{SKU: "8675309"}, &resp)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestCategoriesRetrieval(t *testing.T) {
	Convey("Given a catalog service", t, func() {
		repo := newFakeRepo()
		svc := service.NewCatalogService(repo)
		ctx := context.Background()

		Convey("Querying categories should invoke repository", func() {
			repo.shouldFail = false
			var resp catalog.AllCategoriesResponse
			err := svc.GetProductCategories(ctx, &catalog.AllCategoriesRequest{}, &resp)
			So(err, ShouldBeNil)
			So(len(resp.Categories), ShouldEqual, 2)
			So(resp.Categories[0].Name, ShouldEqual, "Electronics")
		})

		Convey("querying categories should fail when the repo fails", func() {
			repo.shouldFail = true
			var resp catalog.AllCategoriesResponse
			err := svc.GetProductCategories(ctx, &catalog.AllCategoriesRequest{}, &resp)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestProductsWithinCategory(t *testing.T) {
	Convey("Given a catalog service", t, func() {
		repo := newFakeRepo()
		svc := service.NewCatalogService(repo)
		ctx := context.Background()

		Convey("querying products within a category should invoke repository", func() {
			repo.shouldFail = false
			var resp catalog.CategoryProductsResponse
			err := svc.GetProductsInCategory(ctx, &catalog.CategoryProductsRequest{
				CategoryID: 42,
			}, &resp)
			So(err, ShouldBeNil)
			So(len(resp.Products), ShouldEqual, 2)
			So(resp.Products[1].SKU, ShouldEqual, "ABC123")
		})

		Convey("querying products within a non-existent category should produce appropriate error", func() {
			repo.shouldFail = false
			var resp catalog.CategoryProductsResponse
			err := svc.GetProductsInCategory(ctx, &catalog.CategoryProductsRequest{
				CategoryID: 1,
			}, &resp)
			So(err, ShouldBeNil)
			So(resp.Error, ShouldNotBeNil)
			So(resp.Error.HttpHint, ShouldEqual, http.StatusNotFound)
			So(resp.Error.Code, ShouldEqual, catalog.ErrorCode_NOSUCHCATEGORY)
		})

		Convey("querying products for a category should fail when the repo fails", func() {
			repo.shouldFail = true
			var resp catalog.CategoryProductsResponse
			err := svc.GetProductsInCategory(ctx, &catalog.CategoryProductsRequest{
				CategoryID: 42,
			}, &resp)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestProductSearch(t *testing.T) {
	Convey("Given a catalog service", t, func() {
		repo := newFakeRepo()
		svc := service.NewCatalogService(repo)
		ctx := context.Background()

		Convey("search should invoke repository find", func() {
			repo.findCount = 0
			var resp catalog.SearchResponse
			err := svc.ProductSearch(ctx,
				&catalog.SearchRequest{
					SearchTerm: "foo",
					Categories: []uint64{1, 2, 3},
				}, &resp)
			So(err, ShouldBeNil)
			So(repo.findCount, ShouldEqual, 1)
		})

		Convey("search should fail when the catalog repository fails", func() {
			repo.shouldFail = true
			var resp catalog.SearchResponse
			err := svc.ProductSearch(ctx,
				&catalog.SearchRequest{
					SearchTerm: "foo",
					Categories: []uint64{1, 2, 3},
				}, &resp)
			So(err, ShouldNotBeNil)
		})

		Convey("invalid search term should cause a search to fail", func() {
			repo.shouldFail = false
			var resp catalog.SearchResponse
			err := svc.ProductSearch(ctx,
				&catalog.SearchRequest{
					SearchTerm: "",
					Categories: []uint64{1, 2, 3},
				}, &resp)
			So(err, ShouldBeNil)
			So(resp.Error, ShouldNotBeNil)
			So(resp.Error.HttpHint, ShouldEqual, http.StatusBadRequest)
			So(resp.Error.Code, ShouldEqual, catalog.ErrorCode_BADSEARCHREQUEST)
		})
	})
}

type fakeRepo struct {
	shouldFail bool
	findCount  int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{}
}

func (r *fakeRepo) GetProduct(sku string) (product *catalog.Product, err error) {
	if r.shouldFail {
		return nil, errors.New("Faily Fail")
	}

	product = &catalog.Product{
		SKU: sku,
	}
	return
}

func (r *fakeRepo) GetCategories() (categories []*catalog.ProductCategory, err error) {
	if r.shouldFail {
		return nil, errors.New("Faily Fail")
	}
	return []*catalog.ProductCategory{
		&catalog.ProductCategory{CategoryID: 42, Name: "Electronics", Description: "Super electronicy electronics"},
		&catalog.ProductCategory{CategoryID: 12, Name: "Toys", Description: "Toys"},
	}, nil
}

func (r *fakeRepo) GetProductsInCategory(categoryID uint64) (products []*catalog.Product, err error) {
	if r.shouldFail {
		return nil, errors.New("Faily Fail")
	}
	if categoryID == 42 {
		return []*catalog.Product{
			&catalog.Product{SKU: "ABC000"},
			&catalog.Product{SKU: "ABC123"},
		}, nil
	}
	return
}

func (r *fakeRepo) CategoryExists(categoryID uint64) (bool, error) {
	return categoryID == 42, nil
}

func (r *fakeRepo) ProductExists(sku string) (bool, error) {
	return sku == "8675309", nil
}

func (r *fakeRepo) Find(searchTerm string, categories []uint64) (products []*catalog.Product, err error) {
	if r.shouldFail {
		return nil, errors.New("Faily Fail")
	}
	r.findCount++
	return
}
