CREATE DATABASE Person;

USE Person;

--DROP TABLE UserName
CREATE TABLE UserName (
    Id UNIQUEIDENTIFIER NOT NULL PRIMARY KEY DEFAULT(NEWID()),
    FirstName VARCHAR(50) NOT NULL,
    MiddleName VARCHAR(50) NULL,
    LastName VARCHAR(100) NOT NULL,
    Suffix VARCHAR(20),
    CreatedDateTime DATETIME NOT NULL DEFAULT(CURRENT_TIMESTAMP),
    UpdatedDateTime DATETIME NULL
);

insert into username(firstname, lastname)
values ('Bill', 'Gates');

insert into username(firstname, lastname)
values ('Tom', 'Hanks');

insert into username(firstname, middlename, lastname)
values ('John', 'Fitzgerald', 'Quimby');

-- insert into username(firstname, lastname)
-- values ('', '');

select * from UserName
