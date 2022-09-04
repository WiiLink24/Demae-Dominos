from flask_sqlalchemy import SQLAlchemy
import sqlalchemy, json

db = SQLAlchemy()


class DictType(sqlalchemy.types.TypeDecorator):
    impl = sqlalchemy.Text()

    def process_bind_param(self, value, dialect):
        if value is not None:
            value = json.dumps(value)

        return value

    def process_result_value(self, value, dialect):
        if value is not None:
            value = json.loads(value)
        return value


class User(db.Model):
    # For internal use by the channel
    area_code = db.Column(db.String, nullable=False, primary_key=True)
    auth_key = db.Column(db.String)
    basket = db.Column(DictType)
    mac_address = db.Column(db.String)
    order_id = db.Column(db.String)
    price = db.Column(db.String)
