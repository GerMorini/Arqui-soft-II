package repository

import (
	"clase03-memcached/internal/dao"
	"clase03-memcached/internal/domain"
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoItemsRepository implementa ItemsRepository usando DB
type MongoItemsRepository struct {
	col *mongo.Collection // Referencia a la colección "items" en DB
}

// NewMongoItemsRepository crea una nueva instancia del repository
// Recibe una referencia a la base de datos DB
func NewMongoItemsRepository(ctx context.Context, uri, dbName, collectionName string) *MongoItemsRepository {
	opt := options.Client().ApplyURI(uri)
	opt.SetServerSelectionTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
		return nil
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		log.Fatalf("Error pinging DB: %v", err)
		return nil
	}

	return &MongoItemsRepository{
		col: client.Database(dbName).Collection(collectionName), // Conecta con la colección "items"
	}
}

// List obtiene todos los items de DB
func (r *MongoItemsRepository) List(ctx context.Context) ([]domain.Item, error) {
	// ⏰ Timeout para evitar que la operación se cuelgue
	// Esto es importante en producción para no bloquear indefinidamente
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 🔍 Find() sin filtros retorna todos los documentos de la colección
	// bson.M{} es un filtro vacío (equivale a {} en DB shell)
	cur, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx) // ⚠️ IMPORTANTE: Siempre cerrar el cursor para liberar recursos

	// 📦 Decodificar resultados en slice de DAO (modelo DB)
	// Usamos el modelo DAO porque maneja ObjectID y tags BSON
	var daoItems []dao.Item
	if err := cur.All(ctx, &daoItems); err != nil {
		return nil, err
	}

	// 🔄 Convertir de DAO a Domain (para la capa de negocio)
	// Separamos los modelos: DAO para DB, Domain para lógica de negocio
	domainItems := make([]domain.Item, len(daoItems))
	for i, daoItem := range daoItems {
		domainItems[i] = daoItem.ToDomain() // Función definida en dao/Item.go
	}

	return domainItems, nil
}

// Create inserta un nuevo item en DB
// Consigna 1: Validar name y price >= 0, agregar timestamps
func (r *MongoItemsRepository) Create(ctx context.Context, item domain.Item) (domain.Item, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now

	daoItem := dao.FromDomain(item)

	res, err := r.col.InsertOne(ctx, daoItem)
	if err != nil {
		return domain.Item{}, err
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		daoItem.ID = oid
	}

	return daoItem.ToDomain(), nil
}

// GetByID busca un item por su ID
// Consigna 2: Validar que el ID sea un ObjectID válido
func (r *MongoItemsRepository) GetByID(ctx context.Context, id string) (domain.Item, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.Item{}, errors.New("invalid id format")
	}

	var d dao.Item
	err = r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&d)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Item{}, errors.New("item not found")
		}
		return domain.Item{}, err
	}

	return d.ToDomain(), nil
}

// Update actualiza un item existente
// Consigna 3: Update parcial + actualizar updatedAt
func (r *MongoItemsRepository) Update(ctx context.Context, id string, item domain.Item) (domain.Item, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.Item{}, errors.New("invalid id format")
	}

	update := bson.M{
		"$set": bson.M{
			"name":       item.Name,
			"price":      item.Price,
			"updated_at": time.Now().UTC(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated dao.Item
	err = r.col.FindOneAndUpdate(ctx, bson.M{"_id": oid}, update, opts).Decode(&updated)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Item{}, errors.New("item not found")
		}
		return domain.Item{}, err
	}

	return updated.ToDomain(), nil
}

// Delete elimina un item por ID
// Consigna 4: Eliminar documento de DB
func (r *MongoItemsRepository) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	res, err := r.col.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("item not found")
	}
	return nil
}
