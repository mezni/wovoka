from datetime import datetime
from pydantic import BaseModel
from src.domain.types import UUIDType, generate_uuid


class Provider(BaseModel):

    provider_code: UUIDType
    provider_name: str