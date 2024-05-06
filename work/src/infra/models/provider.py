from sqlalchemy.orm import relationship
from sqlalchemy import UUID, Column, String, ForeignKey
from src.infra.base import Base
from src.infra.models.org import OrgModel


class ProviderModel(Base):
    __tablename__ = "providers"
    provider_code = Column(UUID, primary_key=True, index=True)
    provider_name = Column(String(60), unique=True, nullable=False, index=True)
    org_code = Column(UUID, ForeignKey("orgs.org_code"), index=True)
