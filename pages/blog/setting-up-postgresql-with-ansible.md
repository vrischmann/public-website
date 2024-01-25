---
title: Setting up PostgreSQL with Ansible
description: |
    Article explaining how to use Ansible to set up a PostgreSQL server and database on a remote server
date: "2023 July 16"
format: blog_entry
require_prism: true
---

# Introduction

This post will describe in depth how to install and configure PostgreSQL in a single-node deployment using Ansible.

Note that I assume the reader has some basic knowledge of Ansible, Ansible Vault and PostgreSQL.

# What is the goal

I have a small VPS running some services which use PostgreSQL. This server is managed using Ansible; the goal for me was to automate as much as possible the setup of PostgreSQL in that context.

The minimum requirements are:

- installing PostgreSQL
- configuring PostgreSQL
- creating databases, schemas, users and assigning privileges

At the end of the article we should have a minimal playbook that does this.

# Control node setup

## Collection

We will use the [community.postgresql](https://docs.ansible.com/ansible/latest/collections/community/postgresql/index.html) collection to do almost everything in this playbook. This requires having the collection downloaded locally; I recommend using [ansible-galaxy](https://galaxy.ansible.com) to do that.

Create a file called `galaxy-requirements.yml` with this content:

```yaml
collections:
  - community.postgresql
```

Then you can install the collections with this command:

```bash
ansible-galaxy collection install -r galaxy-requirements.yml
```

## Inventory

We need a very basic inventory for this playbook:

```bash
centos ansible_host=192.168.122.30
```

## Variables and vault file

We define variables in the `vars.yml` file. We also need a vault file named `vault.yml` for the user credentials, it can be created like this:

```bash
ansible-vault create vault.yml
```

We will get to these files later.

# Installation

Let’s start with the installation part of the playbook. Since I’m targeting CentOS and Fedora this will use dnf and systemd.

Create a `playbook.yml` with the following content:

```yaml
- hosts: all

  vars_files:
    - vault.yml
    - vars.yml

	handlers:
		- name: Restart postgresql
      ansible.builtin.systemd:
        name: postgresql
        state: restarted

  tasks:
    - name: Install PostgreSQL and psycopg2
      ansible.builtin.dnf:
        name: postgresql-server,postgresql-contrib,python3-psycopg2
        state: present
        update_cache: true

    - name: Create the cluster
      ansible.builtin.command:
        cmd: postgresql-setup --initdb
        creates: /var/lib/pgsql/data/PG_VERSION

    - name: Start and enable the service
      ansible.builtin.systemd:
        name: postgresql
        state: started
        enabled: true
```

A couple of things to note:

- we include the vault and variables using `vars_files`
- we define a “Restart postgresql” handler that will be used later on
- we also install `python3-psycopg2` which is necessary for the `community.postgresql` collection
- we run `postgresql-setup --initdb` if the cluster is not setup
- we start the service only after the cluster has been created, otherwise it will fail

At this point you can verify that PostgreSQL is up and running with `systemctl status postgresql`.

# Configuring system options

Usually we might want to configure a few options in the `postgresql.conf` file. This can be done entirely with the [ALTER SYSTEM](https://www.postgresql.org/docs/15/sql-altersystem.html) SQL command and has completely replaced using a template for my use cases. We will leverage the [community.postgresql.postgresql_set](https://docs.ansible.com/ansible/latest/collections/community/postgresql/postgresql_set_module.html#ansible-collections-community-postgresql-postgresql-set-module) module to do this.

Add this to the `playbook.yml` file:

```yaml
- name: Set options
  community.postgresql.postgresql_set:
    name: "{{ item.name }}"
    value: "{{ item.value }}"
  become: true
  become_user: postgres
  with_items: "{{ postgresql_options }}"
  notify:
    - Restart postgresql
```

The task expects a list of options in the `postgresql_options` variable, let’s define it:

```yaml
postgresql_options:
  - { name: listen_addresses,  value: "localhost, 192.168.122.17" }
  - { name: logging_collector, value: "off"                       }
```

A couple of things to note:

- we act as the `postgres` user so make sure the ansible user you’re using has the ability to `sudo` as `postgres`. This is necessary because the default for the PostgreSQL module is to connect using the Unix-domain socket as the `postgres` user.
- we notify the “Restart postgresql” handler here so that PostgreSQL takes the changes into account.

# Creating databases, schemas and users

Now let’s create our databases, schemas and users.

We will leverage the following modules to do this:
* [community.postgresql.postgresql_db](https://docs.ansible.com/ansible/latest/collections/community/postgresql/postgresql_db_module.html#ansible-collections-community-postgresql-postgresql-db-module)
* [community.postgresql.postgresql_user](https://docs.ansible.com/ansible/latest/collections/community/postgresql/postgresql_user_module.html#ansible-collections-community-postgresql-postgresql-user-module)
* [community.postgresql.postgresql_schema](https://docs.ansible.com/ansible/latest/collections/community/postgresql/postgresql_schema_module.html#ansible-collections-community-postgresql-postgresql-schema-module)

Add this to the `playbook.yml` file:

```yaml
- name: Create the databases
  community.postgresql.postgresql_db:
    name: "{{ item }}"
    encoding: "UTF-8"
  become: true
  become_user: postgres
  with_items: "{{ postgresql_databases }}"

- name: Create the users
  community.postgresql.postgresql_user:
    db: "{{ item.db }}"
    name: "{{ item.name }}"
    password: "{{ item.password | default(omit) }}"
  become: true
  become_user: postgres
  environment:
    PGOPTIONS: "-c password_encryption=scram-sha-256"
  with_items: "{{ postgresql_users }}"

- name: Create the schemas
  community.postgresql.postgresql_schema:
    db: "{{ item.db }}"
    name: "{{ item.name }}"
    owner: "{{ item.name }}"
  become: true
  become_user: postgres
  with_items: "{{ postgresql_schemas }}"
```

One thing to note: we encrypt the user password using `scram-sha-256` (see the [official documentation](https://www.postgresql.org/docs/current/auth-password.html) for an explanation of what it is), this requires setting the environment variable `PGOPTIONS` (see the [notes](https://docs.ansible.com/ansible/latest/collections/community/postgresql/postgresql_user_module.html#notes) of the Ansible module).

Each task expects a list of definitions as input: `postgresql_databases`, `postgresql_users` and `postgresql_schemas`. Let’s define them in the `vars.yml` file:

```yaml
postgresql_databases: [db1, db2]

postgresql_users:
  - { db: db1, name: vincent, password: "{{ vault_postgresql_vincent_password }}" }
  - { db: db2, name: foobar,  password: "{{ vault_postgresql_foobar_password }}"  }

postgresql_schemas:
  - { db: db1, name: vincent, owner: vincent }
  - { db: db2, name: foobar,  owner: foobar  }
```

Pay attention to the `vault_` prefixed variables. These are variables that are defined in the `vault.yml` file. Edit the vault using `ansible-vault edit vault.yml` and add the following:

```yaml
vault_postgresql_vincent_password: "vincent"
vault_postgresql_foobar_password: "foobar"
```

I also encourage you to look at the documentation of each module used here because they can do a lot more; this playbook only reflects what I needed in the past.

# Setting up user privileges

We need to grant privileges to our users, we leverage the [community.postgresql.postgresql_privs](https://docs.ansible.com/ansible/latest/collections/community/postgresql/postgresql_privs_module.html#ansible-collections-community-postgresql-postgresql-privs-module) module to do this.

Add this to the `playbook.yml` file:

```yaml
- name: Set the user privileges
  community.postgresql.postgresql_privs:
    database: "{{ item.db }}"
    state: present
    objs: "{{ item.objs | default(omit) }}"
    privs: "{{ item.privs }}"
    type: "{{ item.type | default(omit) }}"
    roles: "{{ item.roles | default(omit) }}"
  become: true
  become_user: postgres
  with_items: "{{ postgresql_privs | default([]) }}"
```

As before, the task expects a list of privileges in the `postgresql_privs` variable, let’s define it in the `vars.yml` file:

```yaml
postgresql_privs:
  - { db: db1, roles: vincent, privs: ALL, type: database }
  - { db: db2, roles: foobar,  privs: ALL, type: database }
```

This is enough for the users `vincent` and `foobar` to have complete control of the `db1` and `db2` databases respectively.

I won’t explain in detail what the privileges do here, look at the module documentation for an explanation. I encourage you to also look at the [PostgreSQL documentation](https://www.postgresql.org/docs/current/ddl-priv.html) of the privileges DDL.

# Host based authentication

The last missing piece is configuring the *host-based authentication* (the `pg_hba.conf` file).

We leverage the [community.postgresql.postgresql_pg_hba](https://docs.ansible.com/ansible/latest/collections/community/postgresql/postgresql_pg_hba_module.html#ansible-collections-community-postgresql-postgresql-pg-hba-module) module to do this.

Add this to the `playbook.yml` file:

```yaml
- name: Add entries to pg_hba
  community.postgresql.postgresql_pg_hba:
    dest: "/var/lib/pgsql/data/pg_hba.conf"
    address: "{{ item.address | default(omit) }}"
    contype: "{{ item.contype }}"
    databases: "{{ item.databases }}"
    method: "{{ item.method }}"
    users: "{{ item.users }}"
  become: true
  become_user: postgres
  with_items: "{{ postgresql_hba_entries }}"
  notify:
    - Restart postgresql
```

The task expects a list of HBA entries in the `postgresql_hba_entries` variable, let’s define it in the `vars.yml` file:

```yaml
postgresql_hba_entries:
  - { contype: local, databases: all, users: all,                        method: peer          }
  - { contype: host,  databases: db1, users: vincent,  address: samenet, method: scram-sha-256 }
  - { contype: host,  databases: db2, users: foobar,   address: samenet, method: scram-sha-256 }
```

This is enough to connect with either `vincent` or `foobar` using TCP on the same network.

Again, I won’t explain in detail what the entries do, look at the module documentation and the [PostgreSQL documentation](https://www.postgresql.org/docs/current/auth-pg-hba-conf.html) to know more.

# Enabling extensions

Finally, the last piece that may be optional for you but was needed in my use case: enabling an extension for a database.

We leverage the [community.postgresql.postgresql_ext](https://docs.ansible.com/ansible/latest/collections/community/postgresql/postgresql_ext_module.html#ansible-collections-community-postgresql-postgresql-ext-module) module to do this.

Add this to the `playbook.yml` file:

```yaml
- name: Enable the HSTORE extension
  community.postgresql.postgresql_ext:
    name: "{{ item.name }}"
    db: "{{ item.db }}"
    state: present
  become: true
  become_user: postgres
  with_items: "{{ postgresql_extensions }}"
  notify:
    - Restart postgresql
```

The task expects a list of extensions to enable in the `postgresql_extensions` variable, let’s define it in the `vars.yml` file:

```yaml
postgresql_extensions:
  - { db: db1, name: hstore }
```

# Conclusion

This is by no means the *only* way to setup PostgreSQL using Ansible but it works well for my use cases where I configure multiple databases and multiple users. Of course if you don’t need this, you can simply define everything in the playbook instead of in variables.

Let’s put everything together now, this is the final playbook:

```yaml
- hosts: all

  vars_files:
    - vault.yml
    - vars.yml

  handlers:
    - name: Restart postgresql
      ansible.builtin.systemd:
        name: postgresql
        state: restarted

  tasks:
    - name: Install PostgreSQL and psycopg2
      ansible.builtin.dnf:
        name: postgresql-server,postgresql-contrib,python3-psycopg2
        state: present
        update_cache: true

    - name: Create the cluster
      ansible.builtin.command:
        cmd: postgresql-setup --initdb
        creates: /var/lib/pgsql/data/PG_VERSION

    - name: Start and enable the service
      ansible.builtin.systemd:
        name: postgresql
        state: started
        enabled: true

    - name: Set options
      community.postgresql.postgresql_set:
        name: "{{ item.name }}"
        value: "{{ item.value }}"
      become: true
      become_user: postgres
      with_items: "{{ postgresql_options | default([]) }}"
      notify:
        - Restart postgresql

    - name: Create the databases
      community.postgresql.postgresql_db:
        name: "{{ item }}"
        encoding: "UTF-8"
      become: true
      become_user: postgres
      with_items: "{{ postgresql_databases }}"

    - name: Create the users
      community.postgresql.postgresql_user:
        db: "{{ item.db }}"
        name: "{{ item.name }}"
        password: "{{ item.password | default(omit) }}"
      become: true
      become_user: postgres
      environment:
        PGOPTIONS: "-c password_encryption=scram-sha-256"
      with_items: "{{ postgresql_users }}"

    - name: Create the schemas
      community.postgresql.postgresql_schema:
        db: "{{ item.db }}"
        name: "{{ item.name }}"
        owner: "{{ item.name }}"
      become: true
      become_user: postgres
      with_items: "{{ postgresql_schemas }}"

    - name: Set the user privileges
      community.postgresql.postgresql_privs:
        database: "{{ item.db }}"
        state: present
        objs: "{{ item.objs | default(omit) }}"
        privs: "{{ item.privs }}"
        type: "{{ item.type | default(omit) }}"
        roles: "{{ item.roles | default(omit) }}"
      become: true
      become_user: postgres
      with_items: "{{ postgresql_privs | default([]) }}"

    - name: Add entries to pg_hba
      community.postgresql.postgresql_pg_hba:
        dest: "/var/lib/pgsql/data/pg_hba.conf"
        address: "{{ item.address | default(omit) }}"
        contype: "{{ item.contype }}"
        databases: "{{ item.databases }}"
        method: "{{ item.method }}"
        users: "{{ item.users }}"
      become: true
      become_user: postgres
      with_items: "{{ postgresql_hba_entries }}"
      notify:
        - Restart postgresql

    - name: Enable the HSTORE extension
      community.postgresql.postgresql_ext:
        name: "{{ item.name }}"
        db: "{{ item.db }}"
        state: present
      become: true
      become_user: postgres
      with_items: "{{ postgresql_extensions | default([]) }}"
      notify:
        - Restart postgresql
```

This is the final `vars.yml` file:

```yaml
postgresql_options:
  - { name: listen_addresses,  value: "localhost, 192.168.122.17" }
  - { name: logging_collector, value: "off"                       }

postgresql_databases: [db1, db2]

postgresql_users:
  - { db: db1, name: vincent, password: "{{ vault_postgresql_vincent_password }}" }
  - { db: db2, name: foobar,  password: "{{ vault_postgresql_foobar_password }}"  }

postgresql_schemas:
  - { db: db1, name: vincent, owner: vincent }
  - { db: db2, name: foobar,  owner: foobar  }

postgresql_privs:
  - { db: db1, roles: vincent, privs: ALL, type: database }
  - { db: db2, roles: foobar,  privs: ALL, type: database }

postgresql_hba_entries:
  - { contype: local, databases: all, users: all,                        method: peer          }
  - { contype: host,  databases: db1, users: vincent,  address: samenet, method: scram-sha-256 }
  - { contype: host,  databases: db2, users: foobar,   address: samenet, method: scram-sha-256 }

postgresql_extensions:
  - { db: db1, name: hstore }
```

And finally the final `vault.yml` file:

```yaml
vault_postgresql_vincent_password: "vincent"
vault_postgresql_foobar_password: "foobar"
```

I hope you learned something useful! If you have any feedback, feel free to reach out (my contact information is listed at the bottom).
