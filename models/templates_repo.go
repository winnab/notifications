package models

import (
    "database/sql"
    "time"
)

type TemplatesRepoInterface interface {
    Find(ConnectionInterface, string) (Template, error)
    Upsert(ConnectionInterface, Template) (Template, error)
    Destroy(ConnectionInterface, string) error
}

type TemplatesRepo struct{}

func NewTemplatesRepo() TemplatesRepo {
    return TemplatesRepo{}
}

func (repo TemplatesRepo) Find(conn ConnectionInterface, templateName string) (Template, error) {
    template := Template{}
    err := conn.SelectOne(&template, "SELECT * FROM `templates` WHERE `name`=?", templateName)
    if err != nil {
        if err == sql.ErrNoRows {
            return template, ErrRecordNotFound{}
        }
        return template, err
    }
    return template, nil
}

func (repo TemplatesRepo) Upsert(conn ConnectionInterface, template Template) (Template, error) {
    existingTemplate, err := repo.Find(conn, template.Name)
    if err != nil {
        if (err == ErrRecordNotFound{}) {
            return repo.Create(conn, template)
        }
        return Template{}, err
    }

    template.Primary = existingTemplate.Primary
    template.CreatedAt = existingTemplate.CreatedAt
    template.UpdatedAt = time.Now().Truncate(1 * time.Second).UTC()
    _, err = conn.Update(&template)
    if err != nil {
        return Template{}, err
    }

    return template, nil
}

func (repo TemplatesRepo) Create(conn ConnectionInterface, template Template) (Template, error) {
    if (template.CreatedAt == time.Time{}) {
        template.CreatedAt = time.Now().Truncate(1 * time.Second).UTC()
    }
    template.UpdatedAt = template.CreatedAt
    err := conn.Insert(&template)
    if err != nil {
        return Template{}, err
    }
    return template, nil
}

func (repo TemplatesRepo) Destroy(conn ConnectionInterface, templateName string) error {
    template, err := repo.Find(conn, templateName)
    if err != nil {
        if (err == ErrRecordNotFound{}) {
            return nil
        }
        return err
    }

    _, err = conn.Delete(&template)

    return err
}