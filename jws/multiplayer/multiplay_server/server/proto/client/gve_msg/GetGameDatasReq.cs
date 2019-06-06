// automatically generated by the FlatBuffers compiler, do not modify

namespace gve_msg
{

using System;
using FlatBuffers;

public sealed class GetGameDatasReq : Table {
  public static GetGameDatasReq GetRootAsGetGameDatasReq(ByteBuffer _bb) { return GetRootAsGetGameDatasReq(_bb, new GetGameDatasReq()); }
  public static GetGameDatasReq GetRootAsGetGameDatasReq(ByteBuffer _bb, GetGameDatasReq obj) { return (obj.__init(_bb.GetInt(_bb.Position) + _bb.Position, _bb)); }
  public GetGameDatasReq __init(int _i, ByteBuffer _bb) { bb_pos = _i; bb = _bb; return this; }

  public string AccountId { get { int o = __offset(4); return o != 0 ? __string(o + bb_pos) : null; } }
  public ArraySegment<byte>? GetAccountIdBytes() { return __vector_as_arraysegment(4); }
  public string RoomID { get { int o = __offset(6); return o != 0 ? __string(o + bb_pos) : null; } }
  public ArraySegment<byte>? GetRoomIDBytes() { return __vector_as_arraysegment(6); }
  public string Secret { get { int o = __offset(8); return o != 0 ? __string(o + bb_pos) : null; } }
  public ArraySegment<byte>? GetSecretBytes() { return __vector_as_arraysegment(8); }

  public static Offset<GetGameDatasReq> CreateGetGameDatasReq(FlatBufferBuilder builder,
      StringOffset accountIdOffset = default(StringOffset),
      StringOffset roomIDOffset = default(StringOffset),
      StringOffset secretOffset = default(StringOffset)) {
    builder.StartObject(3);
    GetGameDatasReq.AddSecret(builder, secretOffset);
    GetGameDatasReq.AddRoomID(builder, roomIDOffset);
    GetGameDatasReq.AddAccountId(builder, accountIdOffset);
    return GetGameDatasReq.EndGetGameDatasReq(builder);
  }

  public static void StartGetGameDatasReq(FlatBufferBuilder builder) { builder.StartObject(3); }
  public static void AddAccountId(FlatBufferBuilder builder, StringOffset accountIdOffset) { builder.AddOffset(0, accountIdOffset.Value, 0); }
  public static void AddRoomID(FlatBufferBuilder builder, StringOffset roomIDOffset) { builder.AddOffset(1, roomIDOffset.Value, 0); }
  public static void AddSecret(FlatBufferBuilder builder, StringOffset secretOffset) { builder.AddOffset(2, secretOffset.Value, 0); }
  public static Offset<GetGameDatasReq> EndGetGameDatasReq(FlatBufferBuilder builder) {
    int o = builder.EndObject();
    return new Offset<GetGameDatasReq>(o);
  }
  public static void FinishGetGameDatasReqBuffer(FlatBufferBuilder builder, Offset<GetGameDatasReq> offset) { builder.Finish(offset.Value); }
};


}
