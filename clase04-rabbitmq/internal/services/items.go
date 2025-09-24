package services

import (
	"clase04-rabbitmq/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"
)

// ItemsRepository define las operaciones de datos para Items
// Patrón Repository: abstrae el acceso a datos del resto de la aplicación
type ItemsRepository interface {
	// List retorna todos los items de la base de datos
	List(ctx context.Context) ([]domain.Item, error)

	// Create inserta un nuevo item en DB
	Create(ctx context.Context, item domain.Item) (domain.Item, error)

	// GetByID busca un item por su ID
	GetByID(ctx context.Context, id string) (domain.Item, error)

	// Update actualiza un item existente
	Update(ctx context.Context, id string, item domain.Item) (domain.Item, error)

	// Delete elimina un item por ID
	Delete(ctx context.Context, id string) error
} // ItemsServiceImpl implementa ItemsService

type ItemsPublisher interface {
	Publish(ctx context.Context, action string, itemID string) error
}

type ItemsServiceImpl struct {
	repository ItemsRepository // Inyección de dependencia
	cache      ItemsRepository // Inyección de dependencia
	publisher  ItemsPublisher
}

// NewItemsService crea una nueva instancia del service
// Pattern: Dependency Injection - recibe dependencies como parámetros
func NewItemsService(repository ItemsRepository, cache ItemsRepository, publisher ItemsPublisher) ItemsServiceImpl {
	return ItemsServiceImpl{
		repository: repository,
		cache:      cache,
		publisher:  publisher,
	}
}

// List obtiene todos los items
// ✅ IMPLEMENTADO - Delegación simple al repository
func (s *ItemsServiceImpl) List(ctx context.Context) ([]domain.Item, error) {
	// En este caso, no hay lógica de negocio especial
	// Solo delegamos al repository
	return s.repository.List(ctx)
}

// Create valida y crea un nuevo item
// Consigna 1: Validar name no vacío y price >= 0
func (s *ItemsServiceImpl) Create(ctx context.Context, item domain.Item) (domain.Item, error) {
	created, err := s.repository.Create(ctx, item)
	if err != nil {
		return domain.Item{}, fmt.Errorf("error creating item in repository: %w", err)
	}

	if err := s.publisher.Publish(ctx, "create", created.ID); err != nil {
		return domain.Item{}, fmt.Errorf("error publishing item creation: %w", err)
	}

	_, err = s.cache.Create(ctx, created)
	if err != nil {
		return domain.Item{}, fmt.Errorf("error creating item in cache: %w", err)
	}

	return created, nil
}

// GetByID obtiene un item por su ID
// Consigna 2: Validar formato de ID antes de consultar DB
func (s *ItemsServiceImpl) GetByID(ctx context.Context, id string) (domain.Item, error) {
	item, err := s.cache.GetByID(ctx, id)
	if err != nil {
		item, err := s.repository.GetByID(ctx, id)
		if err != nil {
			return domain.Item{}, fmt.Errorf("error getting item from repository: %w", err)
		}

		_, err = s.cache.Create(ctx, item)
		if err != nil {
			return domain.Item{}, fmt.Errorf("error creating item in cache: %w", err)
		}

		return item, nil
	}
	return item, nil
}

// Update actualiza un item existente
// Consigna 3: Validar campos antes de actualizar
func (s *ItemsServiceImpl) Update(ctx context.Context, id string, item domain.Item) (domain.Item, error) {
	// Validar datos de entrada
	if err := s.validateItem(item); err != nil {
		return domain.Item{}, fmt.Errorf("invalid item: %w", err)
	}

	// Actualizar en DB
	updated, err := s.repository.Update(ctx, id, item)
	if err != nil {
		return domain.Item{}, fmt.Errorf("error updating item in repository: %w", err)
	}

	// Publicar evento de actualización (best-effort: si falla, devolver error)
	if err := s.publisher.Publish(ctx, "update", updated.ID); err != nil {
		return domain.Item{}, fmt.Errorf("error publishing item update: %w", err)
	}

	// Guardar en cache (best-effort: si falla, devolver error para aprendizaje)
	if _, err := s.cache.Update(ctx, id, updated); err != nil {
		return domain.Item{}, fmt.Errorf("error updating item in cache: %w", err)
	}

	return updated, nil
}

// Delete elimina un item por ID
// Consigna 4: Validar ID antes de eliminar
func (s *ItemsServiceImpl) Delete(ctx context.Context, id string) error {
	// Borrar de DB primero
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("error deleting item from repository: %w", err)
	}

	// Publicar evento de eliminación
	if err := s.publisher.Publish(ctx, "delete", id); err != nil {
		return fmt.Errorf("error publishing item deletion: %w", err)
	}

	// Borrar de cache
	if err := s.cache.Delete(ctx, id); err != nil {
		return fmt.Errorf("error deleting item from cache: %w", err)
	}

	return nil
}

// validateItem aplica reglas de negocio para validar un item
// 🎯 Función helper para reutilizar validaciones
func (s *ItemsServiceImpl) validateItem(item domain.Item) error {
	// 📝 Name es obligatorio y no puede estar vacío
	if strings.TrimSpace(item.Name) == "" {
		return errors.New("name is required and cannot be empty")
	}

	// 💰 Price debe ser >= 0 (productos gratis están permitidos)
	if item.Price < 0 {
		return errors.New("price must be greater than or equal to 0")
	}

	// ✅ Todas las validaciones pasaron
	return nil
}
