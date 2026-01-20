package router

import (
	backend "server/notes"

	"github.com/gofiber/fiber/v2"
)

func Route(app *fiber.App) {
	app.Post("/createnotes", backend.CreateNotes)
	app.Get("/notes", backend.GetNotes)
	app.Post("/updatenotes", backend.UpdateNotes)
	app.Post("/deletenotes", backend.DeleteNotes)
}
