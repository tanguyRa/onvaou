from fastapi import APIRouter, HTTPException, status


router = APIRouter(prefix="/events", tags=["events"])


@router.get("")
async def list_events():
    raise HTTPException(
        status_code=status.HTTP_501_NOT_IMPLEMENTED,
        detail="Event search is not implemented yet",
    )
