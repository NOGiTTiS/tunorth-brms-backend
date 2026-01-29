package http

import (
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type ResourceHandler struct {
	service ports.ResourceService
}

func NewResourceHandler(service ports.ResourceService) *ResourceHandler {
	return &ResourceHandler{service: service}
}

func (h *ResourceHandler) CreateResource(c *fiber.Ctx) error {
	var res domain.Resource
	if err := c.BodyParser(&res); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	if err := h.service.CreateResource(&res); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *ResourceHandler) GetAllResources(c *fiber.Ctx) error {
	resources, err := h.service.GetAllResources()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(resources)
}

func (h *ResourceHandler) UpdateResource(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	var res domain.Resource
	if err := c.BodyParser(&res); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	if err := h.service.UpdateResource(uint(id), &res); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Resource updated"})
}

func (h *ResourceHandler) DeleteResource(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	if err := h.service.DeleteResource(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Resource deleted"})
}